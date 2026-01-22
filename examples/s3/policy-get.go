//go:build example
// +build example

// Example: Get bucket policy
// Demonstrates how to retrieve a bucket access policy using the RustFS Go SDK
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

	// Get Bucket service
	bucketSvc := client.Bucket()

	bucketName := YOURBUCKET

	// Get bucket policy
	policy, err := bucketSvc.GetPolicy(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to get bucket policy: %v", err)
	}

	log.Printf("✅ Policy for bucket '%s':", bucketName)
	log.Println("----------------------------------------")
	if policy == "" {
		log.Println("Bucket has no policy set")
	} else {
		log.Println(policy)
	}
	log.Println("----------------------------------------")
}
