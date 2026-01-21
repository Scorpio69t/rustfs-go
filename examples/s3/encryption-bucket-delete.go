// Example: Delete bucket default encryption configuration
package main

import (
	"context"
	"fmt"
	"log"

	rustfs "github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/sse"
)

func main() {
	// Connection configuration
	const (
		ACCESS_KEY = "XhJOoEKn3BM6cjD2dVmx"
		SECRET_KEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		ENDPOINT   = "127.0.0.1:9000"
	)

	bucketName := "encrypted-bucket"

	// Create RustFS client
	client, err := rustfs.New(ENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(ACCESS_KEY, SECRET_KEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	ctx := context.Background()
	bucketSvc := client.Bucket()

	// Ensure bucket exists
	exists, err := bucketSvc.Exists(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := bucketSvc.Create(ctx, bucketName); err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
		fmt.Printf("Bucket created: %s\n", bucketName)
	}

	if err := bucketSvc.DeleteEncryption(ctx, bucketName); err != nil {
		log.Fatalf("Failed to delete bucket encryption: %v", err)
	}
	fmt.Printf("Bucket encryption deleted for %s\n", bucketName)

	// Verify deletion
	if _, err := bucketSvc.GetEncryption(ctx, bucketName); err != nil {
		if err == sse.ErrNoEncryptionConfig {
			fmt.Printf("Confirmed: no encryption configuration for %s\n", bucketName)
			return
		}
		log.Fatalf("Failed to verify deletion: %v", err)
	}

	fmt.Printf("Warning: encryption configuration still present for %s\n", bucketName)
}
