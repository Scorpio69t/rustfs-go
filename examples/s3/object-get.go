//go:build example
// +build example

// Example: Download an object
// Demonstrates how to download an object using the RustFS Go SDK
package main

import (
	"context"
	"io"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
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

	// Object to download
	bucketName := YOURBUCKET
	objectName := "my-test-object.txt"

	// Download the object
	reader, objInfo, err := objectSvc.Get(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to download object: %v", err)
	}
	defer reader.Close()

	// Show object info
	log.Println("✅ Object downloaded successfully!")
	log.Printf("   Key: %s", objInfo.Key)
	log.Printf("   Size: %d bytes", objInfo.Size)
	log.Printf("   Content-Type: %s", objInfo.ContentType)
	log.Printf("   ETag: %s", objInfo.ETag)
	log.Printf("   LastModified: %s", objInfo.LastModified.Format("2006-01-02 15:04:05"))

	// Show user metadata (if any)
	if len(objInfo.UserMetadata) > 0 {
		log.Println("   User metadata:")
		for key, value := range objInfo.UserMetadata {
			log.Printf("     %s: %s", key, value)
		}
	}

	// Read content and display
	log.Println("\nObject content:")
	log.Println("----------------------------------------")

	content, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read object content: %v", err)
	}

	log.Println(string(content))
	log.Println("----------------------------------------")
}
