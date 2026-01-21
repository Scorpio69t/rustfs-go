// Debug SSE-C upload behavior
package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"strings"

	rustfs "github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/sse"
)

func main() {
	const (
		ACCESS_KEY = "XhJOoEKn3BM6cjD2dVmx"
		SECRET_KEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		ENDPOINT   = "127.0.0.1:9000"
	)

	client, err := rustfs.New(ENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(ACCESS_KEY, SECRET_KEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	bucketName := "test-encryption"
	objectName := "debug-test.txt"

	// Ensure bucket exists
	bucketSvc := client.Bucket()
	exists, _ := bucketSvc.Exists(ctx, bucketName)
	if !exists {
		bucketSvc.Create(ctx, bucketName)
	}

	// Test 1: Upload without encryption
	fmt.Println("=== Test 1: Upload without encryption ===")
	content1 := "This is test data without encryption"
	reader1 := strings.NewReader(content1)
	size1 := int64(len(content1))

	objectSvc := client.Object()
	info1, err := objectSvc.Put(ctx, bucketName, "no-encryption.txt", reader1, size1)
	if err != nil {
		log.Printf("No-encryption upload failed: %v", err)
	} else {
		fmt.Printf("✓ Upload succeeded, size: %d bytes\n", info1.Size)
	}

	// Test 2: Upload with SSE-C
	fmt.Println("\n=== Test 2: SSE-C encrypted upload ===")

	encKey := make([]byte, 32)
	rand.Read(encKey)
	fmt.Printf("Key: %x...\n", encKey[:8])

	sseC, err := sse.NewSSEC(encKey)
	if err != nil {
		log.Fatalf("Failed to create SSE-C: %v", err)
	}

	content2 := "This is SSE-C encrypted test data"
	fmt.Printf("Content length: %d bytes\n", len(content2))

	// Use bytes.Buffer so the reader can be re-read if necessary
	buffer := bytes.NewBufferString(content2)
	size2 := int64(buffer.Len())
	fmt.Printf("Buffer size: %d bytes\n", size2)

	info2, err := objectSvc.Put(ctx, bucketName, objectName, buffer, size2,
		object.WithSSE(sseC),
	)
	if err != nil {
		log.Fatalf("SSE-C upload failed: %v", err)
	}

	fmt.Printf("✓ Upload returned, size: %d bytes\n", info2.Size)
	fmt.Printf("  ETag: %s\n", info2.ETag)

	// Verify by Stat
	fmt.Println("\n=== Test 3: Download verification ===")
	stat, err := objectSvc.Stat(ctx, bucketName, objectName)
	if err != nil {
		log.Printf("Stat failed: %v", err)
	} else {
		fmt.Printf("Actual object size: %d bytes\n", stat.Size)
	}
}
