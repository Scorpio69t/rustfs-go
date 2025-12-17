package signer

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func TestNewStreamingReader(t *testing.T) {
	data := strings.NewReader("test data")
	reader := io.NopCloser(data)

	streamReader := NewStreamingReader(
		reader,
		"accessKey",
		"secretKey",
		"",
		"us-east-1",
		9, // "test data" 长度
		time.Now(),
		"seed-signature",
	)

	if streamReader == nil {
		t.Fatal("Expected non-nil streaming reader")
	}

	if streamReader.accessKey != "accessKey" {
		t.Errorf("Expected accessKey=accessKey, got %s", streamReader.accessKey)
	}

	if streamReader.contentLen != 9 {
		t.Errorf("Expected contentLen=9, got %d", streamReader.contentLen)
	}
}

func TestStreamingReader_Read(t *testing.T) {
	// 创建测试数据
	testData := "Hello, World!"
	data := strings.NewReader(testData)
	reader := io.NopCloser(data)

	// 创建流式 reader
	streamReader := NewStreamingReader(
		reader,
		"AKIAIOSFODNN7EXAMPLE",
		"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"",
		"us-east-1",
		int64(len(testData)),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		"4f232c4386841ef735655705268965c44a0e4690baa4adea153f7db9fa80a0a9",
	)

	// 读取所有数据
	var buf bytes.Buffer
	_, err := io.Copy(&buf, streamReader)
	if err != nil {
		t.Fatalf("Failed to read from streaming reader: %v", err)
	}

	// 验证输出包含签名分块格式
	output := buf.String()

	// 应该包含 chunk-signature
	if !strings.Contains(output, "chunk-signature=") {
		t.Error("Expected output to contain 'chunk-signature='")
	}

	// 应该包含原始数据
	if !strings.Contains(output, testData) {
		t.Error("Expected output to contain original data")
	}

	// 应该以空分块结束 (0;chunk-signature=...)
	if !strings.Contains(output, "0;chunk-signature=") {
		t.Error("Expected output to contain final empty chunk")
	}

	t.Logf("Streaming output length: %d bytes", buf.Len())
}

func TestStreamingReader_LargeData(t *testing.T) {
	// 创建大于一个分块的数据
	testData := strings.Repeat("A", PayloadChunkSize+1000)
	data := strings.NewReader(testData)
	reader := io.NopCloser(data)

	streamReader := NewStreamingReader(
		reader,
		"AKIAIOSFODNN7EXAMPLE",
		"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"",
		"us-east-1",
		int64(len(testData)),
		time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
		"4f232c4386841ef735655705268965c44a0e4690baa4adea153f7db9fa80a0a9",
	)

	// 读取所有数据
	var buf bytes.Buffer
	_, err := io.Copy(&buf, streamReader)
	if err != nil {
		t.Fatalf("Failed to read from streaming reader: %v", err)
	}

	// 验证输出
	output := buf.String()

	// 应该有多个分块签名
	chunkCount := strings.Count(output, "chunk-signature=")
	if chunkCount < 2 {
		t.Errorf("Expected at least 2 chunks, got %d", chunkCount)
	}

	t.Logf("Streaming output: %d chunks, %d bytes", chunkCount, buf.Len())
}

func TestStreamingReader_EmptyData(t *testing.T) {
	data := strings.NewReader("")
	reader := io.NopCloser(data)

	streamReader := NewStreamingReader(
		reader,
		"accessKey",
		"secretKey",
		"",
		"us-east-1",
		0,
		time.Now(),
		"seed-signature",
	)

	// 读取数据
	var buf bytes.Buffer
	_, err := io.Copy(&buf, streamReader)
	if err != nil {
		t.Fatalf("Failed to read from streaming reader: %v", err)
	}

	// 应该只有一个空分块
	output := buf.String()
	if !strings.Contains(output, "0;chunk-signature=") {
		t.Error("Expected output to contain empty chunk")
	}
}

func TestGetStreamLength(t *testing.T) {
	tests := []struct {
		name      string
		dataLen   int64
		chunkSize int64
		minLen    int64 // 最小预期长度
	}{
		{
			name:      "Empty data",
			dataLen:   0,
			chunkSize: PayloadChunkSize,
			minLen:    0,
		},
		{
			name:      "Small data (less than one chunk)",
			dataLen:   100,
			chunkSize: PayloadChunkSize,
			minLen:    100 + 100, // data + overhead
		},
		{
			name:      "Exactly one chunk",
			dataLen:   PayloadChunkSize,
			chunkSize: PayloadChunkSize,
			minLen:    PayloadChunkSize + 100,
		},
		{
			name:      "Multiple chunks",
			dataLen:   PayloadChunkSize*2 + 1000,
			chunkSize: PayloadChunkSize,
			minLen:    PayloadChunkSize*2 + 1000 + 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			length := GetStreamLength(tt.dataLen, tt.chunkSize)

			if tt.dataLen == 0 {
				if length != 0 {
					t.Errorf("Expected length=0 for empty data, got %d", length)
				}
			} else {
				if length < tt.minLen {
					t.Errorf("Expected length>=%d, got %d", tt.minLen, length)
				}

				// 流式长度应该大于原始数据长度（因为有签名开销）
				if length <= tt.dataLen {
					t.Errorf("Expected stream length (%d) > data length (%d)", length, tt.dataLen)
				}
			}

			t.Logf("Data length: %d, Stream length: %d, Overhead: %d",
				tt.dataLen, length, length-tt.dataLen)
		})
	}
}

func TestBuildChunkStringToSign(t *testing.T) {
	reqTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	streamReader := &StreamingReader{
		region:        "us-east-1",
		reqTime:       reqTime,
		prevSignature: "prev-signature",
	}

	chunkHash := "chunk-hash-value"
	stringToSign := streamReader.buildChunkStringToSign(chunkHash)

	// 验证格式
	lines := strings.Split(stringToSign, "\n")
	if len(lines) != 6 {
		t.Errorf("Expected 6 lines, got %d", len(lines))
	}

	if lines[0] != "AWS4-HMAC-SHA256-PAYLOAD" {
		t.Errorf("Expected first line to be algorithm, got %s", lines[0])
	}

	if lines[3] != "prev-signature" {
		t.Errorf("Expected previous signature on line 4, got %s", lines[3])
	}

	if lines[4] != EmptySHA256 {
		t.Errorf("Expected empty SHA256 on line 5, got %s", lines[4])
	}

	if lines[5] != chunkHash {
		t.Errorf("Expected chunk hash on line 6, got %s", lines[5])
	}

	t.Logf("String to sign:\n%s", stringToSign)
}

func TestStreamingReader_Close(t *testing.T) {
	data := strings.NewReader("test")
	reader := io.NopCloser(data)

	streamReader := NewStreamingReader(
		reader,
		"accessKey",
		"secretKey",
		"",
		"us-east-1",
		4,
		time.Now(),
		"seed-signature",
	)

	err := streamReader.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}
}

func BenchmarkStreamingReader(b *testing.B) {
	testData := strings.Repeat("A", PayloadChunkSize)
	reqTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data := strings.NewReader(testData)
		reader := io.NopCloser(data)

		streamReader := NewStreamingReader(
			reader,
			"AKIAIOSFODNN7EXAMPLE",
			"wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			"",
			"us-east-1",
			int64(len(testData)),
			reqTime,
			"4f232c4386841ef735655705268965c44a0e4690baa4adea153f7db9fa80a0a9",
		)

		io.Copy(io.Discard, streamReader)
		streamReader.Close()
	}
}

func BenchmarkGetStreamLength(b *testing.B) {
	dataLen := int64(PayloadChunkSize * 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetStreamLength(dataLen, PayloadChunkSize)
	}
}
