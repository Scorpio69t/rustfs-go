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

// 流式签名相关常量
const (
	// StreamingSignAlgorithm AWS 流式签名算法
	StreamingSignAlgorithm = "STREAMING-AWS4-HMAC-SHA256-PAYLOAD"

	// PayloadChunkSize 默认分块大小 (64KB)
	PayloadChunkSize = 64 * 1024

	// EmptySHA256 空内容的 SHA256 哈希
	EmptySHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

// StreamingReader 实现分块上传签名的 io.Reader
type StreamingReader struct {
	// 访问密钥
	accessKey    string
	secretKey    string
	sessionToken string

	// 区域和时间
	region  string
	reqTime time.Time

	// 签名
	prevSignature string
	seedSignature string

	// 底层 reader
	baseReader io.ReadCloser

	// 内容长度
	contentLen int64
	bytesRead  int64

	// 缓冲区
	buf      bytes.Buffer
	chunkBuf []byte
	chunkLen int

	// 状态
	done bool
}

// NewStreamingReader 创建新的流式签名 reader
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

// Read 实现 io.Reader 接口
func (s *StreamingReader) Read(p []byte) (n int, err error) {
	// 如果已完成且缓冲区为空，返回 EOF
	if s.done && s.buf.Len() == 0 {
		return 0, io.EOF
	}

	// 如果缓冲区有数据，先从缓冲区读取
	if s.buf.Len() > 0 {
		return s.buf.Read(p)
	}

	// 读取下一个分块
	if err := s.readNextChunk(); err != nil {
		return 0, err
	}

	// 从缓冲区读取
	return s.buf.Read(p)
}

// readNextChunk 读取并签名下一个分块
func (s *StreamingReader) readNextChunk() error {
	if s.done {
		return io.EOF
	}

	// 读取数据
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

		// 如果读取了足够的数据或已读取所有数据，跳出循环
		if s.chunkLen >= PayloadChunkSize || s.bytesRead >= s.contentLen {
			break
		}
	}

	// 验证读取的字节数
	if s.done && s.bytesRead != s.contentLen {
		return fmt.Errorf("content length mismatch: expected %d, got %d", s.contentLen, s.bytesRead)
	}

	// 签名分块
	s.signChunk(s.chunkLen)

	// 如果已完成，写入最后的空分块
	if s.done {
		s.signChunk(0)
	}

	return nil
}

// signChunk 签名一个分块
func (s *StreamingReader) signChunk(chunkLen int) {
	// 计算分块的 SHA256
	h := sha256.New()
	if chunkLen > 0 {
		h.Write(s.chunkBuf[:chunkLen])
	}
	chunkHash := hex.EncodeToString(h.Sum(nil))

	// 构建待签名字符串
	stringToSign := s.buildChunkStringToSign(chunkHash)

	// 计算签名
	signingKey := streamingDeriveSigningKey(s.secretKey, s.region, s.reqTime)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	// 更新前一个签名
	s.prevSignature = signature

	// 写入分块头部：<chunk-size-hex>;chunk-signature=<signature>\r\n
	chunkHeader := fmt.Sprintf("%x;chunk-signature=%s\r\n", chunkLen, signature)
	s.buf.WriteString(chunkHeader)

	// 写入分块数据
	if chunkLen > 0 {
		s.buf.Write(s.chunkBuf[:chunkLen])
	}

	// 写入分块尾部：\r\n
	s.buf.WriteString("\r\n")
}

// buildChunkStringToSign 构建分块的待签名字符串
func (s *StreamingReader) buildChunkStringToSign(chunkHash string) string {
	// 格式：
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

// streamingBuildCredentialScope 构建凭证范围
func streamingBuildCredentialScope(region string, t time.Time) string {
	return fmt.Sprintf("%s/%s/%s/aws4_request",
		t.Format("20060102"),
		region,
		serviceTypeS3)
}

// streamingDeriveSigningKey 派生签名密钥
func streamingDeriveSigningKey(secretKey, region string, t time.Time) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secretKey), []byte(t.Format("20060102")))
	regionKey := hmacSHA256(dateKey, []byte(region))
	serviceKey := hmacSHA256(regionKey, []byte(serviceTypeS3))
	signingKey := hmacSHA256(serviceKey, []byte("aws4_request"))
	return signingKey
}

// Close 关闭底层 reader
func (s *StreamingReader) Close() error {
	return s.baseReader.Close()
}

// GetStreamLength 计算流式签名的总长度
func GetStreamLength(dataLen int64, chunkSize int64) int64 {
	if dataLen <= 0 {
		return 0
	}

	if chunkSize <= 0 {
		chunkSize = PayloadChunkSize
	}

	// 计算分块数量
	numChunks := dataLen / chunkSize
	remainder := dataLen % chunkSize

	streamLen := int64(0)

	// 完整分块的长度
	// 每个分块格式：<hex-size>;chunk-signature=<64-char-sig>\r\n<data>\r\n
	// hex-size: 最多 16 字符 (64KB = 0x10000)
	// ";chunk-signature=" : 17 字符
	// signature: 64 字符
	// \r\n: 2 字符
	// data: chunkSize 字节
	// \r\n: 2 字符
	for i := int64(0); i < numChunks; i++ {
		hexSize := fmt.Sprintf("%x", chunkSize)
		streamLen += int64(len(hexSize)) + 17 + 64 + 2 + chunkSize + 2
	}

	// 最后的不完整分块
	if remainder > 0 {
		hexSize := fmt.Sprintf("%x", remainder)
		streamLen += int64(len(hexSize)) + 17 + 64 + 2 + remainder + 2
	}

	// 最后的空分块 (表示结束)
	// 格式：0;chunk-signature=<64-char-sig>\r\n\r\n
	streamLen += 1 + 17 + 64 + 2 + 2

	return streamLen
}

// PrepareStreamingRequest 准备流式签名请求
func PrepareStreamingRequest(req *http.Request, sessionToken string, dataLen int64) {
	// 设置必要的头部
	req.Header.Set("X-Amz-Content-Sha256", StreamingSignAlgorithm)
	req.Header.Set("Content-Encoding", "aws-chunked")
	req.Header.Set("X-Amz-Decoded-Content-Length", strconv.FormatInt(dataLen, 10))

	if sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", sessionToken)
	}

	// 设置内容长度为流式签名后的长度
	req.ContentLength = GetStreamLength(dataLen, PayloadChunkSize)
}
