// Example: Get bucket default encryption configuration
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
		ACCESS_KEY = "rustfsadmin"
		SECRET_KEY = "rustfsadmin"
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

	config, err := bucketSvc.GetEncryption(ctx, bucketName)
	if err != nil {
		if err == sse.ErrNoEncryptionConfig {
			fmt.Printf("No encryption configuration found for %s\n", bucketName)
			return
		}
		log.Fatalf("Failed to get bucket encryption: %v", err)
	}

	fmt.Printf("Bucket encryption configuration for %s:\n", bucketName)
	for i, rule := range config.Rules {
		fmt.Printf("  Rule %d:\n", i+1)
		fmt.Printf("    Algorithm: %s\n", rule.ApplySSEByDefault.SSEAlgorithm)
		if rule.ApplySSEByDefault.KMSMasterKeyID != "" {
			fmt.Printf("    KMS Key ID: %s\n", rule.ApplySSEByDefault.KMSMasterKeyID)
		}
		fmt.Printf("    BucketKeyEnabled: %v\n", rule.BucketKeyEnabled)
	}
}
