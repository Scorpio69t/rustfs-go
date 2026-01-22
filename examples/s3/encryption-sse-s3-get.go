// Example: Download an object encrypted with SSE-S3
//
// SSE-S3 uses server-managed keys (AES-256). No extra headers are required on GET.
package main

import (
	"context"
	"fmt"
	"io"
	"log"

	rustfs "github.com/Scorpio69t/rustfs-go"
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
	objectSvc := client.Object()

	// Download object (server will decrypt automatically)
	reader, objInfo, err := objectSvc.Get(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to download object. Upload it first with encryption-sse-s3-put.go. Error: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read object: %v", err)
	}

	fmt.Printf("? Downloaded SSE-S3 object\n")
	fmt.Printf("  Bucket: %s\n", objInfo.Bucket)
	fmt.Printf("  Object: %s\n", objInfo.Key)
	fmt.Printf("  ETag: %s\n", objInfo.ETag)
	fmt.Printf("  Size: %d bytes\n", objInfo.Size)
	fmt.Printf("  Content: %s\n", string(data))
}
