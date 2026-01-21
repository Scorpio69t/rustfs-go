//go:build example
// +build example

// Example: Delete a bucket
// Demonstrates how to delete a bucket using the RustFS Go SDK
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

	// Name of the bucket to delete
	bucketName := "test-bucket-to-delete"

	// Delete the bucket
	// Note: bucket must be empty to be deleted
	err = bucketSvc.Delete(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to delete bucket: %v", err)
	}

	log.Printf("✅ Bucket deleted: %s", bucketName)
	log.Println("\nNote: Only empty buckets can be deleted")
	log.Println("If the bucket contains objects, delete them first")
}
