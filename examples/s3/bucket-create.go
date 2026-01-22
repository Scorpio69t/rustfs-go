//go:build example
// +build example

// Example: Create a bucket
//
// Demonstrates how to create a new bucket using the RustFS Go SDK.
package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/bucket"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		ACCESS_KEY = "rustfsadmin"
		SECRET_KEY = "rustfsadmin"
		ENDPOINT   = "127.0.0.1:9000"
		BUCKET     = "mybucket"
	)

	// Initialize RustFS client
	client, err := rustfs.New(ENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(ACCESS_KEY, SECRET_KEY, ""),
		Secure:      false, // Use HTTP for local testing, set to true in production
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Get Bucket service
	bucketSvc := client.Bucket()

	// Bucket to create
	bucketName := BUCKET

	// Create bucket with options (set region)
	err = bucketSvc.Create(ctx, bucketName,
		bucket.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatalf("Failed to create bucket: %v", err)
	}

	log.Printf("✅ Bucket created: %s", bucketName)

	// Verify bucket exists
	exists, err := bucketSvc.Exists(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}

	if exists {
		log.Printf("✅ Bucket '%s' exists", bucketName)
	}
}
