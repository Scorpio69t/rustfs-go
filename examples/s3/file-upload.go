//go:build example
// +build example

// Example: Upload an object from a file
//
// Demonstrates how to upload a local file as an object using the RustFS Go SDK.
package main

import (
	"context"
	"log"
	"os"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		ACCESS_KEY = "XhJOoEKn3BM6cjD2dVmx"
		SECRET_KEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		ENDPOINT   = "127.0.0.1:9000"
		BUCKET     = "mybucket"
	)

	// Initialize RustFS client
	client, err := rustfs.New(ENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(ACCESS_KEY, SECRET_KEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Get Object service
	objectSvc := client.Object()

	// Upload parameters
	bucketName := BUCKET
	objectName := "uploaded-file.txt"
	filePath := "test-file.txt" // local file to upload

	// Create a test file if it does not exist
	if err := createTestFile(filePath); err != nil {
		log.Fatalf("Failed to create test file: %v", err)
	}

	// Upload file to object
	// FPut detects file size and content type automatically
	uploadInfo, err := objectSvc.FPut(ctx, bucketName, objectName, filePath,
		object.WithContentType("text/plain"),
		object.WithUserMetadata(map[string]string{
			"source": "local-file",
		}),
	)
	if err != nil {
		log.Fatalf("File upload failed: %v", err)
	}

	// Print upload result
	log.Println("✅ File uploaded successfully!")
	log.Printf("   Local file: %s", filePath)
	log.Printf("   Bucket: %s", uploadInfo.Bucket)
	log.Printf("   Object: %s", uploadInfo.Key)
	log.Printf("   ETag: %s", uploadInfo.ETag)
	log.Printf("   Size: %d bytes", uploadInfo.Size)

	if uploadInfo.VersionID != "" {
		log.Printf("   VersionID: %s", uploadInfo.VersionID)
	}

	log.Println("\nTip: use file-download.go to download this object to a file")
}

// createTestFile creates a test file
func createTestFile(filePath string) error {
	// If file exists, reuse it
	if _, err := os.Stat(filePath); err == nil {
		log.Printf("Using existing file: %s", filePath)
		return nil
	}

	// Create new file
	content := "This is a test file for upload demonstration.\n" +
		"RustFS Go SDK - File Upload Example\n" +
		"Generated automatically for testing purposes.\n"

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return err
	}

	log.Printf("Created test file: %s", filePath)
	return nil
}
