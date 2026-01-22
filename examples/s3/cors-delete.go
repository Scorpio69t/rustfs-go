//go:build example
// +build example

// Example: Delete bucket CORS configuration
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
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
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

	if err := bucketSvc.DeleteCORS(ctx, YOURBUCKET); err != nil {
		log.Fatalf("Failed to delete CORS: %v", err)
	}
	fmt.Printf("CORS configuration deleted for %s\n", YOURBUCKET)

	// Verify deletion
	if _, err := bucketSvc.GetCORS(ctx, YOURBUCKET); err != nil {
		if err == cors.ErrNoCORSConfig {
			fmt.Printf("Confirmed: no CORS configuration for %s\n", YOURBUCKET)
			return
		}
		log.Fatalf("Failed to verify deletion: %v", err)
	}

	fmt.Printf("Warning: CORS configuration still present for %s\n", YOURBUCKET)
}
