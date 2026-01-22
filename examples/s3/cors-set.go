//go:build example
// +build example

// Example: Set bucket CORS configuration
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/cors"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "cors-bucket"
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

	// Ensure bucket exists
	exists, err := bucketSvc.Exists(ctx, YOURBUCKET)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := bucketSvc.Create(ctx, YOURBUCKET); err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
		fmt.Printf("Bucket created: %s\n", YOURBUCKET)
	}

	config := cors.NewConfig([]cors.Rule{
		{
			ID:            "example-cors",
			AllowedOrigin: []string{"*"},
			AllowedMethod: []string{"GET", "PUT"},
			AllowedHeader: []string{"*"},
			ExposeHeader:  []string{"x-amz-request-id"},
			MaxAgeSeconds: 3000,
		},
	})

	if err := bucketSvc.SetCORS(ctx, YOURBUCKET, config); err != nil {
		log.Fatalf("Failed to set CORS: %v", err)
	}

	fmt.Printf("CORS configuration set for %s\n", YOURBUCKET)
}
