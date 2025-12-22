// Package signer internal/signer/streaming.go
package signer

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Streaming signing related constants
const (
	// StreamingSignAlgorithm AWS streaming signing algorithm
	StreamingSignAlgorithm = "STREAMING-AWS4-HMAC-SHA256-PAYLOAD"

	// PayloadChunkSize default chunk size (64KB)
	PayloadChunkSize = 64 * 1024

	// EmptySHA256 SHA256 hash of empty content
	EmptySHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

// StreamingReader implements chunked upload signing io.Reader
type StreamingReader struct {
	// Access keys
	accessKey    string
	secretKey    string
	sessionToken string

	// Region and time
	region  string
	reqTime time.Time

	// Signatures
	prevSignature string
	seedSignature string

	// Underlying reader
	baseReader io.ReadCloser

	// Content length
	contentLen int64
	bytesRead  int64

	// Buffers
	buf      bytes.Buffer
	chunkBuf []byte
	chunkLen int

	// State
	done bool
}

// NewStreamingReader creates a new streaming signing reader
func NewStreamingReader(
	reader io.ReadCloser,
	accessKey, secretKey, sessionToken, region string,
	contentLen int64,
	reqTime time.Time,
	seedSignature string,
) *StreamingReader {
	return &StreamingReader{
		accessKey:     accessKey,
		secretKey:     secretKey,
		sessionToken:  sessionToken,
		region:        region,
		reqTime:       reqTime,
		baseReader:    reader,
		contentLen:    contentLen,
		chunkBuf:      make([]byte, PayloadChunkSize),
		seedSignature: seedSignature,
		prevSignature: seedSignature,
	}
}

// Read implements io.Reader
func (s *StreamingReader) Read(p []byte) (n int, err error) {
	// If finished and buffer empty, return EOF
	if s.done && s.buf.Len() == 0 {
		return 0, io.EOF
	}

	// If buffer has data, read from it first
	if s.buf.Len() > 0 {
		return s.buf.Read(p)
	}

	// Read next chunk
	if err := s.readNextChunk(); err != nil {
		return 0, err
	}

	// Read from buffer
	return s.buf.Read(p)
}

// readNextChunk reads and signs next chunk
func (s *StreamingReader) readNextChunk() error {
	if s.done {
		return io.EOF
	}

	// Read data
	s.chunkLen = 0
	for s.chunkLen < PayloadChunkSize {
		n, err := s.baseReader.Read(s.chunkBuf[s.chunkLen:])
		if n > 0 {
			s.chunkLen += n
			s.bytesRead += int64(n)
		}

		if err != nil {
			if err == io.EOF {
				s.done = true
				break
			}
			return err
		}

		// Break if chunk full or all data read
		if s.chunkLen >= PayloadChunkSize || s.bytesRead >= s.contentLen {
			break
		}
	}

	// Validate bytes read
	if s.done && s.bytesRead != s.contentLen {
		return fmt.Errorf("content length mismatch: expected %d, got %d", s.contentLen, s.bytesRead)
	}

	// Sign chunk
	s.signChunk(s.chunkLen)

	// If finished, write final empty chunk
	if s.done {
		s.signChunk(0)
	}

	return nil
}

// signChunk signs a chunk
func (s *StreamingReader) signChunk(chunkLen int) {
	// Compute SHA256 of the chunk
	h := sha256.New()
	if chunkLen > 0 {
		h.Write(s.chunkBuf[:chunkLen])
	}
	chunkHash := hex.EncodeToString(h.Sum(nil))

	// Build string to sign
	stringToSign := s.buildChunkStringToSign(chunkHash)

	// Calculate signature
	signingKey := streamingDeriveSigningKey(s.secretKey, s.region, s.reqTime)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	// Update previous signature
	s.prevSignature = signature

	// Write chunk header: <chunk-size-hex>;chunk-signature=<signature>\r\n
	chunkHeader := fmt.Sprintf("%x;chunk-signature=%s\r\n", chunkLen, signature)
	s.buf.WriteString(chunkHeader)

	// Write chunk data
	if chunkLen > 0 {
		s.buf.Write(s.chunkBuf[:chunkLen])
	}

	// Write chunk trailer: \r\n
	s.buf.WriteString("\r\n")
}

// buildChunkStringToSign builds the string to sign for a chunk
func (s *StreamingReader) buildChunkStringToSign(chunkHash string) string {
	// Format:
	// AWS4-HMAC-SHA256-PAYLOAD\n
	// <timestamp>\n
	// <scope>\n
	// <previous-signature>\n
	// <empty-string-sha256>\n
	// <chunk-sha256>

	scope := streamingBuildCredentialScope(s.region, s.reqTime)

	parts := []string{
		"AWS4-HMAC-SHA256-PAYLOAD",
		s.reqTime.Format(iso8601DateFormat),
		scope,
		s.prevSignature,
		EmptySHA256,
		chunkHash,
	}

	return strings.Join(parts, "\n")
}

// streamingBuildCredentialScope builds credential scope
func streamingBuildCredentialScope(region string, t time.Time) string {
	return fmt.Sprintf("%s/%s/%s/aws4_request",
		t.Format("20060102"),
		region,
		serviceTypeS3)
}

// streamingDeriveSigningKey derives signing key
func streamingDeriveSigningKey(secretKey, region string, t time.Time) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secretKey), []byte(t.Format("20060102")))
	regionKey := hmacSHA256(dateKey, []byte(region))
	serviceKey := hmacSHA256(regionKey, []byte(serviceTypeS3))
	signingKey := hmacSHA256(serviceKey, []byte("aws4_request"))
	return signingKey
}

// Close closes underlying reader
func (s *StreamingReader) Close() error {
	return s.baseReader.Close()
}

// GetStreamLength calculates total length after streaming signature
func GetStreamLength(dataLen int64, chunkSize int64) int64 {
	if dataLen <= 0 {
		return 0
	}

	if chunkSize <= 0 {
		chunkSize = PayloadChunkSize
	}

	// Calculate number of chunks
	numChunks := dataLen / chunkSize
	remainder := dataLen % chunkSize

	streamLen := int64(0)

	// Length of full chunks
	// Each chunk format: <hex-size>;chunk-signature=<64-char-sig>\r\n<data>\r\n
	// hex-size: up to 16 chars (64KB = 0x10000)
	// ";chunk-signature=" : 17 chars
	// signature: 64 chars
	// \r\n: 2 chars
	// data: chunkSize bytes
	// \r\n: 2 chars
	for i := int64(0); i < numChunks; i++ {
		hexSize := fmt.Sprintf("%x", chunkSize)
		streamLen += int64(len(hexSize)) + 17 + 64 + 2 + chunkSize + 2
	}

	// Final incomplete chunk
	if remainder > 0 {
		hexSize := fmt.Sprintf("%x", remainder)
		streamLen += int64(len(hexSize)) + 17 + 64 + 2 + remainder + 2
	}

	// Final empty chunk (end marker)
	// Format: 0;chunk-signature=<64-char-sig>\r\n\r\n
	streamLen += 1 + 17 + 64 + 2 + 2

	return streamLen
}

// PrepareStreamingRequest prepares streaming signed request
func PrepareStreamingRequest(req *http.Request, sessionToken string, dataLen int64) {
	// Set required headers
	req.Header.Set("X-Amz-Content-Sha256", StreamingSignAlgorithm)
	req.Header.Set("Content-Encoding", "aws-chunked")
	req.Header.Set("X-Amz-Decoded-Content-Length", strconv.FormatInt(dataLen, 10))

	if sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", sessionToken)
	}

	// Set content length to streaming-signed length
	req.ContentLength = GetStreamLength(dataLen, PayloadChunkSize)
}
