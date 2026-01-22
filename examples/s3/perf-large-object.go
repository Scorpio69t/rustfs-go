//go:build example
// +build example

// Example: Large object performance
// Measures upload and download throughput for a large object.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/bucket"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

type patternReader struct {
	remaining int64
	pattern   byte
}

func (r *patternReader) Read(p []byte) (int, error) {
	if r.remaining <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	for i := range p {
		p[i] = r.pattern
	}
	r.remaining -= int64(len(p))
	return len(p), nil
}

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
	)

	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	bucketSvc := client.Bucket()
	objectSvc := client.Object()

	bucketName := fmt.Sprintf("perf-large-%d", time.Now().Unix())
	if err := bucketSvc.Create(ctx, bucketName, bucket.WithRegion("us-east-1")); err != nil {
		log.Fatalf("Failed to create bucket: %v", err)
	}
	defer func() {
		if err := bucketSvc.Delete(ctx, bucketName); err != nil {
			log.Printf("Cleanup bucket failed: %v", err)
		}
	}()

	objectName := "large-object.bin"
	sizeBytes := int64(32 * 1024 * 1024) // 32 MB

	start := time.Now()
	reader := &patternReader{remaining: sizeBytes, pattern: 'L'}
	if _, err := objectSvc.Put(
		ctx,
		bucketName,
		objectName,
		reader,
		sizeBytes,
		object.WithContentType("application/octet-stream"),
	); err != nil {
		log.Fatalf("Large upload failed: %v", err)
	}
	uploadDuration := time.Since(start)
	uploadThroughput := float64(sizeBytes) / uploadDuration.Seconds() / (1024 * 1024)
	log.Printf("Upload: %s (%.2f MB/s)", uploadDuration, uploadThroughput)

	start = time.Now()
	getReader, _, err := objectSvc.Get(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Large download failed: %v", err)
	}
	if _, err := io.Copy(io.Discard, getReader); err != nil {
		log.Fatalf("Failed to read large object: %v", err)
	}
	if err := getReader.Close(); err != nil {
		log.Fatalf("Failed to close download reader: %v", err)
	}
	downloadDuration := time.Since(start)
	downloadThroughput := float64(sizeBytes) / downloadDuration.Seconds() / (1024 * 1024)
	log.Printf("Download: %s (%.2f MB/s)", downloadDuration, downloadThroughput)

	if err := objectSvc.Delete(ctx, bucketName, objectName); err != nil {
		log.Fatalf("Failed to delete large object: %v", err)
	}

	log.Println("Large object performance test completed.")
}
