//go:build example
// +build example

// Example: Set bucket tags
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
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "tagging-bucket"
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
	bucketSvc := client.Bucket()

	exists, err := bucketSvc.Exists(ctx, YOURBUCKET)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := bucketSvc.Create(ctx, YOURBUCKET, bucket.WithRegion("us-east-1")); err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
	}

	tags := map[string]string{
		"env":  "dev",
		"team": "storage",
	}

	if err := bucketSvc.SetTagging(ctx, YOURBUCKET, tags); err != nil {
		log.Fatalf("Failed to set bucket tags: %v", err)
	}

	log.Printf("Bucket tags updated for %s", YOURBUCKET)
}
