//go:build example
// +build example

// Example: Download an object with client-side encryption (CSE).
package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
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
		objectKey = "cse-get.txt"
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

	objectSvc := client.Object()
	payload := "CSE example payload"
	_, err = objectSvc.Put(ctx, bucket, objectKey, strings.NewReader(payload), int64(len(payload)),
		object.WithPutCSE(cseClient),
		object.WithContentType("text/plain; charset=utf-8"),
	)
	if err != nil {
		log.Fatalf("Failed to upload encrypted object: %v", err)
	}

	reader, _, err := objectSvc.Get(ctx, bucket, objectKey,
		object.WithGetCSE(cseClient),
	)
	if err != nil {
		log.Fatalf("Failed to download encrypted object: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read object: %v", err)
	}

	fmt.Printf("✓ Downloaded and decrypted content: %s\n", string(data))
}
