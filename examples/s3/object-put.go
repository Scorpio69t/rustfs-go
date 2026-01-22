//go:build example
// +build example

// Example: Upload an object
// Demonstrates how to upload an object to a bucket using the RustFS Go SDK
package main

import (
	"context"
	"log"
	"strings"

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
		YOURBUCKET          = "mybucket"
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

	// Get Object service
	objectSvc := client.Object()

	// Prepare upload data
	bucketName := YOURBUCKET
	objectName := "my-test-object.txt"
	content := "Hello, RustFS! This is a test object uploaded using the RustFS Go SDK."

	// Create reader
	reader := strings.NewReader(content)

	// Upload object
	// Set content type and user metadata via options
	uploadInfo, err := objectSvc.Put(ctx, bucketName, objectName, reader, int64(len(content)),
		object.WithContentType("text/plain; charset=utf-8"),
		object.WithUserMetadata(map[string]string{
			"author":      "rustfs-go-sdk",
			"description": "Example object",
		}),
	)
	if err != nil {
		log.Fatalf("Failed to upload object: %v", err)
	}

	// Show upload result
	log.Println("✅ Object uploaded successfully!")
	log.Printf("   Bucket: %s", uploadInfo.Bucket)
	log.Printf("   Key: %s", uploadInfo.Key)
	log.Printf("   ETag: %s", uploadInfo.ETag)
	log.Printf("   Size: %d bytes", uploadInfo.Size)

	if uploadInfo.VersionID != "" {
		log.Printf("   VersionID: %s", uploadInfo.VersionID)
	}

	log.Println("\nTip: use object-get.go to download this object")
}
