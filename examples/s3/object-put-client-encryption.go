//go:build example
// +build example

// Example: Upload an object with client-side encryption (CSE).
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/cse"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	const (
		accessKey = "rustfsadmin"
		secretKey = "rustfsadmin"
		endpoint  = "127.0.0.1:9000"
		bucket    = "cse-demo"
	)

	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	ctx := context.Background()
	bucketSvc := client.Bucket()
	exists, err := bucketSvc.Exists(ctx, bucket)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := bucketSvc.Create(ctx, bucket); err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
		fmt.Printf("✓ Bucket created: %s\n", bucket)
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatalf("Failed to generate CSE key: %v", err)
	}
	cseClient, err := cse.New(key)
	if err != nil {
		log.Fatalf("Failed to create CSE client: %v", err)
	}
	fmt.Printf("✓ Generated CSE key (keep it safe)\n")

	content := "Client-side encrypted payload from RustFS Go SDK."
	reader := strings.NewReader(content)

	objectSvc := client.Object()
	uploadInfo, err := objectSvc.Put(ctx, bucket, "cse-put.txt", reader, int64(len(content)),
		object.WithPutCSE(cseClient),
		object.WithContentType("text/plain; charset=utf-8"),
	)
	if err != nil {
		log.Fatalf("Failed to upload object: %v", err)
	}

	fmt.Printf("✓ Uploaded encrypted object\n")
	fmt.Printf("  Bucket: %s\n", uploadInfo.Bucket)
	fmt.Printf("  Object: %s\n", uploadInfo.Key)
	fmt.Printf("  Size: %d bytes\n", uploadInfo.Size)
}
