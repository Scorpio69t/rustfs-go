//go:build example
// +build example

// Example: Upload object with progress
// Demonstrates how to display upload progress
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

const (
	endpoint  = "127.0.0.1:9000"
	accessKey = "rustfsadmin"
	secretKey = "rustfsadmin"
	bucket    = "mybucket"
)

// ProgressReader wraps an io.Reader and shows read progress
type ProgressReader struct {
	reader      io.Reader
	total       int64
	current     int64
	lastPercent int64
}

// NewProgressReader creates a new progress reader
func NewProgressReader(r io.Reader, total int64) *ProgressReader {
	return &ProgressReader{
		reader: r,
		total:  total,
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.current += int64(n)

	// Calculate current progress percentage
	if pr.total > 0 {
		percent := pr.current * 100 / pr.total
		// Print every 5% change
		if percent >= pr.lastPercent+5 || err == io.EOF {
			fmt.Printf("\rUpload progress: %d%% (%d/%d bytes)", percent, pr.current, pr.total)
			pr.lastPercent = percent
			if err == io.EOF || percent >= 100 {
				fmt.Println()
			}
		}
	}

	return n, err
}

func main() {
	// Create client
	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	service := client.Object()

	objectName := "progress-upload.txt"

	// Create test data (~5MB)
	data := strings.Repeat("This is a progress display test.", 100000)
	dataSize := int64(len(data))

	fmt.Printf("Preparing to upload object '%s' (size: %.2f MB)...\n", objectName, float64(dataSize)/1024/1024)

	// 使用进度读取器包装数据
	reader := strings.NewReader(data)
	progressReader := NewProgressReader(reader, dataSize)

	// 上传对象
	uploadInfo, err := service.Put(
		ctx,
		bucket,
		objectName,
		progressReader,
		dataSize,
		object.WithContentType("text/plain; charset=utf-8"),
	)
	if err != nil {
		log.Fatalf("\nUpload failed: %v\n", err)
	}

	fmt.Println("\n✅ Upload successful")
	fmt.Printf("Object: %s\n", uploadInfo.Key)
	fmt.Printf("ETag: %s\n", uploadInfo.ETag)
	fmt.Printf("Size: %d bytes (%.2f MB)\n", uploadInfo.Size, float64(uploadInfo.Size)/1024/1024)
}
