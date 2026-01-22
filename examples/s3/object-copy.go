//go:build example
// +build example

// Example: Copy an object
// Demonstrates how to copy objects using the RustFS Go SDK
package main

import (
	"context"
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

	// Source object
	srcBucket := YOURBUCKET
	srcObject := "my-test-object.txt"

	// Destination object
	destBucket := YOURBUCKET
	destObject := "my-test-object-copy.txt"

	// Copy the object
	copyInfo, err := objectSvc.Copy(ctx,
		destBucket, destObject, // 目标
		srcBucket, srcObject, // 源
	)
	if err != nil {
		log.Fatalf("Failed to copy object: %v", err)
	}

	log.Println("✅ Object copied successfully!")
	log.Printf("   Source: %s/%s", srcBucket, srcObject)
	log.Printf("   Destination: %s/%s", destBucket, destObject)
	log.Printf("   ETag: %s", copyInfo.ETag)

	if copyInfo.VersionID != "" {
		log.Printf("   VersionID: %s", copyInfo.VersionID)
	}

	// Example: copy and replace metadata
	log.Println("\n=== Copy and replace metadata ===")
	destObject2 := "my-test-object-copy-with-metadata.txt"

	copyInfo2, err := objectSvc.Copy(ctx,
		destBucket, destObject2,
		srcBucket, srcObject,
		object.WithCopyMetadata(map[string]string{
			"copied-at": "2026-01-20",
			"author":    "rustfs-sdk",
		}, true), // true => replace metadata
	)
	if err != nil {
		log.Printf("Failed to copy object: %v", err)
	} else {
		log.Printf("✅ Copied: %s", destObject2)
		log.Printf("   ETag: %s", copyInfo2.ETag)
	}
}
