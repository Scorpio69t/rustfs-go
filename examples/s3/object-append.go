//go:build example
// +build example

// Example: Append data to an existing object
// Demonstrates the RustFS append API (not part of standard S3).
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

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "append-bucket"
	)

	// Initialize RustFS client
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Ensure bucket exists
	bucketSvc := client.Bucket()
	exists, err := bucketSvc.Exists(ctx, YOURBUCKET)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := bucketSvc.Create(ctx, YOURBUCKET); err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
		fmt.Printf("Bucket created: %s\n", YOURBUCKET)
	}

	objectSvc := client.Object()
	objectName := "append-object.txt"

	part1 := "hello"
	part2 := ", rustfs!"

	// Append first part at offset 0
	info1, err := objectSvc.Append(
		ctx,
		YOURBUCKET,
		objectName,
		strings.NewReader(part1),
		int64(len(part1)),
		0,
		object.WithContentType("text/plain; charset=utf-8"),
	)
	if err != nil {
		log.Fatalf("Failed to append part 1: %v", err)
	}
	fmt.Printf("Appended part 1, size: %d bytes\n", info1.Size)

	// Append second part at the end (offset -1 means auto)
	info2, err := objectSvc.Append(
		ctx,
		YOURBUCKET,
		objectName,
		strings.NewReader(part2),
		int64(len(part2)),
		-1,
	)
	if err != nil {
		log.Fatalf("Failed to append part 2: %v", err)
	}
	fmt.Printf("Appended part 2, size: %d bytes\n", info2.Size)

	// Verify final content
	reader, _, err := objectSvc.Get(ctx, YOURBUCKET, objectName)
	if err != nil {
		log.Fatalf("Failed to download object: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read object: %v", err)
	}

	fmt.Printf("Final content: %s\n", string(data))
}
