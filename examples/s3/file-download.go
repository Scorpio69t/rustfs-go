//go:build example
// +build example

// Example: Download an object to a file
// Demonstrates how to download an object to a local file using the RustFS Go SDK
package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
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

	// Download parameters
	bucketName := YOURBUCKET
	objectName := "my-test-object.txt"
	localFilePath := "downloaded-file.txt"

	// Download object to a file
	objInfo, err := objectSvc.FGet(ctx, bucketName, objectName, localFilePath)
	if err != nil {
		log.Fatalf("Failed to download file: %v", err)
	}

	// Show download result
	log.Println("✅ File downloaded successfully!")
	log.Printf("   Object: %s", objInfo.Key)
	log.Printf("   Saved to: %s", localFilePath)
	log.Printf("   Size: %d bytes", objInfo.Size)
	log.Printf("   Content-Type: %s", objInfo.ContentType)
	log.Printf("   ETag: %s", objInfo.ETag)
	log.Printf("   LastModified: %s", objInfo.LastModified.Format("2006-01-02 15:04:05"))
}
