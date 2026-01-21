//go:build example
// +build example

// Example: Get bucket location
// Demonstrates how to retrieve the region/location of a bucket using the RustFS Go SDK
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

	// Get bucket location
	bucketName := YOURBUCKET
	location, err := bucketSvc.GetLocation(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to get bucket location: %v", err)
	}
	log.Printf("✅ Bucket '%s' is located in region: %s", bucketName, location)

	if location == "" {
		log.Println("Tip: an empty string typically indicates the default region (us-east-1)")
	}
}
