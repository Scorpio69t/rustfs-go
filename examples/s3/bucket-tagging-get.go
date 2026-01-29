//go:build example
// +build example

// Example: Get bucket tags
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

	tags, err := bucketSvc.GetTagging(ctx, YOURBUCKET)
	if err != nil {
		log.Fatalf("Failed to get bucket tags: %v", err)
	}

	if len(tags) == 0 {
		log.Printf("No tags found for bucket %s", YOURBUCKET)
		return
	}

	log.Printf("Bucket tags for %s:", YOURBUCKET)
	for k, v := range tags {
		log.Printf("  %s=%s", k, v)
	}
}
