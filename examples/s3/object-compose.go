//go:build example
// +build example

// Example: Compose a new object from multiple source objects
// Demonstrates the RustFS compose API (server-side multipart copy).
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "compose-bucket"
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
	bucketSvc := client.Bucket()
	objectSvc := client.Object()

	// Ensure bucket exists
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

	// Compose requires all but the last part to be >= 5MiB
	const part1Size = 5 * 1024 * 1024
	part1 := bytes.Repeat([]byte("A"), part1Size)
	part2 := []byte("tail")

	srcObject1 := "compose-part-1.bin"
	srcObject2 := "compose-part-2.bin"
	dstObject := "compose-result.bin"

	if _, err := objectSvc.Put(ctx, YOURBUCKET, srcObject1, bytes.NewReader(part1), int64(len(part1))); err != nil {
		log.Fatalf("Failed to upload part 1: %v", err)
	}
	if _, err := objectSvc.Put(ctx, YOURBUCKET, srcObject2, bytes.NewReader(part2), int64(len(part2))); err != nil {
		log.Fatalf("Failed to upload part 2: %v", err)
	}

	sources := []object.SourceInfo{
		{Bucket: YOURBUCKET, Object: srcObject1},
		{Bucket: YOURBUCKET, Object: srcObject2},
	}
	dst := object.DestinationInfo{Bucket: YOURBUCKET, Object: dstObject}

	uploadInfo, err := objectSvc.Compose(ctx, dst, sources)
	if err != nil {
		log.Fatalf("Failed to compose object: %v", err)
	}

	fmt.Printf("Composed object: %s/%s\n", uploadInfo.Bucket, uploadInfo.Key)
	fmt.Printf("ETag: %s\n", uploadInfo.ETag)
	fmt.Printf("Size: %d bytes\n", uploadInfo.Size)

	// Download and verify size
	reader, _, err := objectSvc.Get(ctx, YOURBUCKET, dstObject)
	if err != nil {
		log.Fatalf("Failed to download composed object: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read composed object: %v", err)
	}

	expectedSize := len(part1) + len(part2)
	fmt.Printf("Downloaded size: %d bytes (expected %d)\n", len(data), expectedSize)
}
