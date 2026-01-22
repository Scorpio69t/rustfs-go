//go:build example
// +build example

// Example: Get object metadata
// Demonstrates how to retrieve object metadata using the RustFS Go SDK
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

	// Object to query
	bucketName := YOURBUCKET
	objectName := "my-test-object.txt"

	// Get object info
	objInfo, err := objectSvc.Stat(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to stat object: %v", err)
	}

	// Show object info
	log.Println("✅ Object info:")
	log.Println("----------------------------------------")
	log.Printf("  Key: %s", objInfo.Key)
	log.Printf("  Bucket: %s", bucketName)
	log.Printf("  Size: %d bytes", objInfo.Size)
	log.Printf("  Content-Type: %s", objInfo.ContentType)
	log.Printf("  ETag: %s", objInfo.ETag)
	log.Printf("  LastModified: %s", objInfo.LastModified.Format("2006-01-02 15:04:05"))

	if objInfo.VersionID != "" {
		log.Printf("  VersionID: %s", objInfo.VersionID)
	}

	// Show user metadata
	if len(objInfo.UserMetadata) > 0 {
		log.Println("\n  User metadata:")
		for key, value := range objInfo.UserMetadata {
			log.Printf("    %s: %s", key, value)
		}
	}

	// Show tags
	if len(objInfo.UserTags) > 0 {
		log.Println("\n  Object tags:")
		for key, value := range objInfo.UserTags {
			log.Printf("    %s: %s", key, value)
		}
	}

	log.Println("----------------------------------------")
}
