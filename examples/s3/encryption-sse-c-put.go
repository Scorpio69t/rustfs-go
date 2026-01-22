// Example: Upload and download object with SSE-C (customer-provided key)
//
// SSE-C uses a 256-bit encryption key provided by the client.
// The key is never stored on the server. The same key must be provided for every access.
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"strings"

	rustfs "github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/sse"
)

func main() {
	// Connection config
	const (
		ACCESS_KEY = "rustfsadmin"
		SECRET_KEY = "rustfsadmin"
		ENDPOINT   = "127.0.0.1:9000"
	)

	bucketName := "test-encryption"
	objectName := "encrypted-with-customer-key.txt"

	// Create RustFS client
	client, err := rustfs.New(ENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(ACCESS_KEY, SECRET_KEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	ctx := context.Background()

	// Create bucket if not exists
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

	// Generate 256-bit (32 bytes) encryption key
	encryptionKey := make([]byte, 32)
	if _, err := rand.Read(encryptionKey); err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}
	fmt.Printf("✓ Generated 256-bit encryption key: %x...\n", encryptionKey[:8])

	// Create SSE-C encrypter
	sseEncrypter, err := sse.NewSSEC(encryptionKey)
	if err != nil {
		log.Fatalf("Failed to create SSE-C encrypter: %v", err)
	}

	// Prepare upload data
	content := "This is highly sensitive data encrypted with a client-provided key. The key is never stored on the server."
	reader := strings.NewReader(content)
	size := int64(len(content))

	// Upload object with SSE-C
	objectSvc := client.Object()
	uploadInfo, err := objectSvc.Put(ctx, bucketName, objectName, reader, size,
		object.WithSSE(sseEncrypter),
		object.WithContentType("text/plain; charset=utf-8"),
	)
	if err != nil {
		log.Fatalf("Failed to upload object: %v", err)
	}

	fmt.Printf("\n✓ Uploaded with SSE-C encryption\n")
	fmt.Printf("  Bucket: %s\n", uploadInfo.Bucket)
	fmt.Printf("  Object: %s\n", uploadInfo.Key)
	fmt.Printf("  ETag: %s\n", uploadInfo.ETag)
	fmt.Printf("  Size: %d bytes\n", uploadInfo.Size)

	// Download object (must provide the same key)
	fmt.Printf("\n📥 Downloading object with the same key...\n")
	downloadReader, _, err := objectSvc.Get(ctx, bucketName, objectName,
		object.WithGetSSE(sseEncrypter),
	)
	if err != nil {
		log.Fatalf("Failed to download object: %v", err)
	}
	defer downloadReader.Close()

	// Read all content (avoid blocking)
	data, err := io.ReadAll(downloadReader)
	if err != nil {
		log.Fatalf("Failed to read object: %v", err)
	}

	fmt.Printf("✓ Downloaded and decrypted with client key\n")
	fmt.Printf("  Content: %s\n", string(data))

	// Test: download with wrong key should fail
	fmt.Printf("\n🔒 Test: download with wrong key...\n")
	wrongKey := make([]byte, 32)
	rand.Read(wrongKey)
	wrongEncrypter, _ := sse.NewSSEC(wrongKey)

	_, _, err = objectSvc.Get(ctx, bucketName, objectName,
		object.WithGetSSE(wrongEncrypter),
	)
	if err != nil {
		fmt.Printf("✓ Correct: download with wrong key failed\n")
		fmt.Printf("  Error: %v\n", err)
	} else {
		fmt.Printf("⚠️  Warning: download with wrong key succeeded (should not happen)\n")
	}

	fmt.Printf("\n📌 SSE-C Notes:\n")
	fmt.Printf("  ✓ Key must be 256 bits (32 bytes)\n")
	fmt.Printf("  ✓ Key is never stored on the server\n")
	fmt.Printf("  ✓ You must provide the same key for every access\n")
	fmt.Printf("  ✓ Losing the key means permanent data loss\n")
	fmt.Printf("  ✓ Suitable for scenarios requiring full key control\n")
	fmt.Printf("  ⚠️  Client must manage keys securely (use a key management system)\n")

	// Cleanup (optional)
	// err = objectSvc.Delete(ctx, bucketName, objectName)
	// if err != nil {
	//  log.Printf("Warning: failed to delete object: %v", err)
	// }
}
