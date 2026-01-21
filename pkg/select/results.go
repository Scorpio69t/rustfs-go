// Package select provides SelectObjectContent helpers.
package s3select

import (
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"net/http"
	"strings"

	"github.com/Scorpio69t/rustfs-go/errors"
)

type preludeInfo struct {
	totalLen  uint32
	headerLen uint32
}

// Results provides a streaming reader for select output.
type Results struct {
	pipeReader *io.PipeReader
	resp       *http.Response
	stats      *StatsMessage
	progress   *ProgressMessage
}

// ProgressMessage contains progress XML messages.
type ProgressMessage struct {
	XMLName xml.Name `xml:"Progress"`
	StatsMessage
}

// StatsMessage contains select stats.
type StatsMessage struct {
	XMLName        xml.Name `xml:"Stats"`
	BytesScanned   int64
	BytesProcessed int64
	BytesReturned  int64
}

type messageType string

const (
	errorMsg  messageType = "error"
	commonMsg messageType = "event"
)

type eventType string

const (
	endEvent      eventType = "End"
	recordsEvent  eventType = "Records"
	progressEvent eventType = "Progress"
	statsEvent    eventType = "Stats"
)

type contentType string

const (
	xmlContent contentType = "text/xml"
)

// NewResults creates a select result parser and starts streaming.
func NewResults(resp *http.Response, bucketName, objectName string) (*Results, error) {
	if resp.StatusCode != http.StatusOK {
		defer closeResponse(resp)
		return nil, errors.ParseErrorResponse(resp, bucketName, objectName)
	}

	pipeReader, pipeWriter := io.Pipe()
	streamer := &Results{
		resp:       resp,
		stats:      &StatsMessage{},
		progress:   &ProgressMessage{},
		pipeReader: pipeReader,
	}
	streamer.start(pipeWriter)
	return streamer, nil
}

// Close closes the response body and stream reader.
func (s *Results) Close() error {
	defer closeResponse(s.resp)
	return s.pipeReader.Close()
}

// Read reads select records from the stream.
func (s *Results) Read(b []byte) (n int, err error) {
	return s.pipeReader.Read(b)
}

// Stats returns select stats after completion.
func (s *Results) Stats() *StatsMessage {
	return s.stats
}

// Progress returns select progress updates.
func (s *Results) Progress() *ProgressMessage {
	return s.progress
}

func (s *Results) start(pipeWriter *io.PipeWriter) {
	go func() {
		for {
			headers := make(http.Header)

			crc := crc32.New(crc32.IEEETable)
			crcReader := io.TeeReader(s.resp.Body, crc)

			prelude, err := processPrelude(crcReader, crc)
			if err != nil {
				pipeWriter.CloseWithError(err)
				closeResponse(s.resp)
				return
			}

			if prelude.headerLen > 0 {
				if err := extractHeader(io.LimitReader(crcReader, int64(prelude.headerLen)), headers); err != nil {
					pipeWriter.CloseWithError(err)
					closeResponse(s.resp)
					return
				}
			}

			payloadLen := prelude.payloadLen()
			switch messageType(headers.Get("message-type")) {
			case errorMsg:
				pipeWriter.CloseWithError(fmt.Errorf("%s:\"%s\"", headers.Get("error-code"), headers.Get("error-message")))
				closeResponse(s.resp)
				return
			case commonMsg:
				switch eventType(headers.Get("event-type")) {
				case endEvent:
					pipeWriter.Close()
					closeResponse(s.resp)
					return
				case recordsEvent:
					if _, err := io.Copy(pipeWriter, io.LimitReader(crcReader, payloadLen)); err != nil {
						pipeWriter.CloseWithError(err)
						closeResponse(s.resp)
						return
					}
				case progressEvent:
					if contentType(headers.Get("content-type")) != xmlContent {
						pipeWriter.CloseWithError(fmt.Errorf("unexpected content-type %s for progress event", headers.Get("content-type")))
						closeResponse(s.resp)
						return
					}
					if err := decodeXML(io.LimitReader(crcReader, payloadLen), s.progress); err != nil {
						pipeWriter.CloseWithError(err)
						closeResponse(s.resp)
						return
					}
				case statsEvent:
					if contentType(headers.Get("content-type")) != xmlContent {
						pipeWriter.CloseWithError(fmt.Errorf("unexpected content-type %s for stats event", headers.Get("content-type")))
						closeResponse(s.resp)
						return
					}
					if err := decodeXML(io.LimitReader(crcReader, payloadLen), s.stats); err != nil {
						pipeWriter.CloseWithError(err)
						closeResponse(s.resp)
						return
					}
				}
			}

			if err := checkCRC(s.resp.Body, crc.Sum32()); err != nil {
				pipeWriter.CloseWithError(err)
				closeResponse(s.resp)
				return
			}
		}
	}()
}

func (p preludeInfo) payloadLen() int64 {
	return int64(p.totalLen - p.headerLen - 16)
}

func processPrelude(prelude io.Reader, crc hash.Hash32) (preludeInfo, error) {
	var err error
	pInfo := preludeInfo{}

	pInfo.totalLen, err = extractUint32(prelude)
	if err != nil {
		return pInfo, err
	}

	pInfo.headerLen, err = extractUint32(prelude)
	if err != nil {
		return pInfo, err
	}

	preCRC := crc.Sum32()
	if err := checkCRC(prelude, preCRC); err != nil {
		return pInfo, err
	}

	return pInfo, nil
}

func extractHeader(body io.Reader, headers http.Header) error {
	for {
		headerName, err := extractHeaderType(body)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if _, err := extractUint8(body); err != nil {
			return err
		}

		headerValue, err := extractHeaderValue(body)
		if err != nil {
			return err
		}

		headers.Set(headerName, headerValue)
	}
	return nil
}

func extractHeaderType(body io.Reader) (string, error) {
	headerNameLen, err := extractUint8(body)
	if err != nil {
		return "", err
	}
	headerName, err := extractString(body, int(headerNameLen))
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(headerName, ":"), nil
}

func extractHeaderValue(body io.Reader) (string, error) {
	bodyLen, err := extractUint16(body)
	if err != nil {
		return "", err
	}
	bodyName, err := extractString(body, int(bodyLen))
	if err != nil {
		return "", err
	}
	return bodyName, nil
}

func extractString(source io.Reader, lenBytes int) (string, error) {
	buf := make([]byte, lenBytes)
	if _, err := readFull(source, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func extractUint32(r io.Reader) (uint32, error) {
	buf := make([]byte, 4)
	if _, err := readFull(r, buf); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(buf), nil
}

func extractUint16(r io.Reader) (uint16, error) {
	buf := make([]byte, 2)
	if _, err := readFull(r, buf); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(buf), nil
}

func extractUint8(r io.Reader) (uint8, error) {
	buf := make([]byte, 1)
	if _, err := readFull(r, buf); err != nil {
		return 0, err
	}
	return buf[0], nil
}

func checkCRC(r io.Reader, expect uint32) error {
	msgCRC, err := extractUint32(r)
	if err != nil {
		return err
	}

	if msgCRC != expect {
		return fmt.Errorf("checksum mismatch: got 0x%X, want 0x%X", msgCRC, expect)
	}
	return nil
}

func decodeXML(r io.Reader, v interface{}) error {
	return xml.NewDecoder(r).Decode(v)
}

func readFull(r io.Reader, buf []byte) (int, error) {
	return io.ReadFull(r, buf)
}

func closeResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}
