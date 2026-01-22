// Example: Download an object encrypted with SSE-C (customer-provided key)
//
// SSE-C requires the same 256-bit key used during upload.
package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"log"

	rustfs "github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
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

	// Use the same key you used for upload (32 bytes, 64 hex chars).
	const KEY_HEX = "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f"

	bucketName := "test-encryption"
	objectName := "encrypted-with-customer-key.txt"

	// Decode the 256-bit key
	key, err := hex.DecodeString(KEY_HEX)
	if err != nil {
		log.Fatalf("Failed to decode key: %v", err)
	}
	if len(key) != 32 {
		log.Fatalf("Key must be 32 bytes, got %d", len(key))
	}

	sseEncrypter, err := sse.NewSSEC(key)
	if err != nil {
		log.Fatalf("Failed to create SSE-C encrypter: %v", err)
	}

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

	// Download object with SSE-C key
	reader, objInfo, err := objectSvc.Get(ctx, bucketName, objectName,
		object.WithGetSSE(sseEncrypter),
	)
	if err != nil {
		log.Fatalf("Failed to download object. Upload it first with encryption-sse-c-put.go. Error: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read object: %v", err)
	}

	fmt.Printf("? Downloaded SSE-C object\n")
	fmt.Printf("  Bucket: %s\n", objInfo.Bucket)
	fmt.Printf("  Object: %s\n", objInfo.Key)
	fmt.Printf("  ETag: %s\n", objInfo.ETag)
	fmt.Printf("  Size: %d bytes\n", objInfo.Size)
	fmt.Printf("  Content: %s\n", string(data))
}
