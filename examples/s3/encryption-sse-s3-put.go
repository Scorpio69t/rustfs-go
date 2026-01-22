// Example: Upload an object using SSE-S3 server-side encryption
//
// SSE-S3 uses server-managed keys (AES-256). The client does not manage keys.
// This is the simplest server-side encryption mode and fits most use cases.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	rustfs "github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		ACCESS_KEY = "rustfsadmin"
		SECRET_KEY = "rustfsadmin"
		ENDPOINT   = "127.0.0.1:9000"
	)

	bucketName := "test-encryption"
	objectName := "encrypted-object.txt"

	// Create RustFS client
	client, err := rustfs.New(ENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(ACCESS_KEY, SECRET_KEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	ctx := context.Background()

	// Create bucket if it does not exist
	bucketSvc := client.Bucket()
	exists, err := bucketSvc.Exists(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		err = bucketSvc.Create(ctx, bucketName)
		if err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
		fmt.Printf("✓ Bucket created: %s\n", bucketName)
	}

	// Prepare upload data
	content := "This is data encrypted by SSE-S3"
	reader := strings.NewReader(content)
	size := int64(len(content))

	// Upload object with SSE-S3
	objectSvc := client.Object()
	uploadInfo, err := objectSvc.Put(ctx, bucketName, objectName, reader, size,
		object.WithSSES3(),
		object.WithContentType("text/plain; charset=utf-8"),
	)
	if err != nil {
		log.Fatalf("Failed to upload object: %v", err)
	}

	fmt.Printf("✓ Uploaded with SSE-S3 encryption\n")
	fmt.Printf("  Bucket: %s\n", uploadInfo.Bucket)
	fmt.Printf("  Object: %s\n", uploadInfo.Key)
	fmt.Printf("  ETag: %s\n", uploadInfo.ETag)
	fmt.Printf("  Size: %d bytes\n", uploadInfo.Size)

	// Download object (server will decrypt automatically)
	downloadReader, _, err := objectSvc.Get(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to download object: %v", err)
	}
	defer downloadReader.Close()

	// Read entire content safely
	data, err := io.ReadAll(downloadReader)
	if err != nil {
		log.Fatalf("Failed to read object: %v", err)
	}

	fmt.Printf("\n✓ Download successful (server-decrypted)\n")
	fmt.Printf("  Content: %s\n", string(data))
}
