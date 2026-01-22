//go:build example
// +build example

// Example: List objects (V2 API)
// Demonstrates how to list objects in a bucket using the RustFS Go SDK
package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// 配置连接参数
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "mybucket"
	)

	// 初始化 RustFS 客户端
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Get Object service
	objectSvc := client.Object()

	bucketName := YOURBUCKET

	// List all objects
	log.Printf("=== Listing all objects in bucket '%s' ===\n", bucketName)
	objectsCh := objectSvc.List(ctx, bucketName)

	count := 0
	for obj := range objectsCh {
		if obj.Err != nil {
			log.Fatalf("Failed to list objects: %v", obj.Err)
		}
		count++

		if obj.IsPrefix {
			log.Printf("%d. %s (prefix)", count, obj.Key)
		} else {
			log.Printf("%d. %s", count, obj.Key)
			log.Printf("   Size: %d bytes", obj.Size)
			log.Printf("   LastModified: %s", obj.LastModified.Format("2006-01-02 15:04:05"))
			log.Printf("   ETag: %s", obj.ETag)
			log.Println()
		}
	}

	if count == 0 {
		log.Println("Bucket is empty")
	} else {
		log.Printf("\nFound %d objects", count)
	}

	// Example: list by prefix
	log.Println("\n=== Listing objects with prefix 'my-test' ===")
	objectsCh = objectSvc.List(ctx, bucketName,
		object.WithListPrefix("my-test"),
	)

	count = 0
	for obj := range objectsCh {
		if obj.Err != nil {
			log.Printf("List objects error: %v", obj.Err)
			break
		}
		count++
		log.Printf("%d. %s (%d bytes)", count, obj.Key, obj.Size)
	}

	if count == 0 {
		log.Println("No matching objects found")
	}
}
