//go:build example
// +build example

// Example: Check if a bucket exists
// Demonstrates how to check bucket existence using the RustFS Go SDK
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

	// Get Bucket service
	bucketSvc := client.Bucket()

	// Check if bucket exists
	bucketName := YOURBUCKET
	exists, err := bucketSvc.Exists(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}

	if exists {
		log.Printf("✅ Bucket '%s' exists", bucketName)
	} else {
		log.Printf("❌ Bucket '%s' does not exist", bucketName)
	}

	// Check a non-existent bucket
	nonExistentBucket := "this-bucket-does-not-exist-12345"
	exists, err = bucketSvc.Exists(ctx, nonExistentBucket)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}

	if exists {
		log.Printf("✅ Bucket '%s' exists", nonExistentBucket)
	} else {
		log.Printf("❌ Bucket '%s' does not exist", nonExistentBucket)
	}
}
