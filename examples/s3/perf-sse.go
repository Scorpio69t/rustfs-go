//go:build example
// +build example

// Example: SSE performance
// Compares upload latency with and without SSE-S3.
package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/bucket"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

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

	bucketName := fmt.Sprintf("perf-sse-%d", time.Now().Unix())
	if err := bucketSvc.Create(ctx, bucketName, bucket.WithRegion("us-east-1")); err != nil {
		log.Fatalf("Failed to create bucket: %v", err)
	}
	defer func() {
		if err := bucketSvc.Delete(ctx, bucketName); err != nil {
			log.Printf("Cleanup bucket failed: %v", err)
		}
	}()

	payloadSize := 5 * 1024 * 1024 // 5 MB
	payload := bytes.Repeat([]byte("a"), payloadSize)

	plainObject := "plain-upload.bin"
	start := time.Now()
	if _, err := objectSvc.Put(
		ctx,
		bucketName,
		plainObject,
		bytes.NewReader(payload),
		int64(len(payload)),
		object.WithContentType("application/octet-stream"),
	); err != nil {
		log.Fatalf("Plain upload failed: %v", err)
	}
	plainDuration := time.Since(start)
	log.Printf("Plain upload: %s", plainDuration)

	sseObject := "sse-upload.bin"
	start = time.Now()
	if _, err := objectSvc.Put(
		ctx,
		bucketName,
		sseObject,
		bytes.NewReader(payload),
		int64(len(payload)),
		object.WithContentType("application/octet-stream"),
		object.WithSSES3(),
	); err != nil {
		log.Fatalf("SSE-S3 upload failed: %v", err)
	}
	sseDuration := time.Since(start)
	log.Printf("SSE-S3 upload: %s", sseDuration)

	if err := objectSvc.Delete(ctx, bucketName, plainObject); err != nil {
		log.Fatalf("Failed to delete plain object: %v", err)
	}
	if err := objectSvc.Delete(ctx, bucketName, sseObject); err != nil {
		log.Fatalf("Failed to delete SSE object: %v", err)
	}

	log.Println("SSE performance comparison completed.")
}
