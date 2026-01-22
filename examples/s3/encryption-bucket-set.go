// Example: Set bucket default encryption (SSE-S3)
//
// After setting default encryption, new objects are encrypted automatically.
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

	// Create bucket if absent
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

	config := sse.NewConfiguration()
	algo := config.Rules[0].ApplySSEByDefault.SSEAlgorithm
	fmt.Printf("Setting default encryption to %s...\n", algo)

	if err := bucketSvc.SetEncryption(ctx, bucketName, *config); err != nil {
		log.Fatalf("Failed to set bucket encryption: %v", err)
	}

	fmt.Printf("Bucket default encryption set for %s\n", bucketName)
}
