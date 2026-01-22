//go:build example
// +build example

// Example: Get object tagging
// Demonstrates how to get tags for an object using the RustFS Go SDK
package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
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

	// 获取 Object 服务
	objectSvc := client.Object()

	// Object info
	bucketName := YOURBUCKET
	objectName := "my-test-object.txt"

	// 获取对象标签
	tags, err := objectSvc.GetTagging(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to get object tagging: %v", err)
	}

	log.Println("✅ Object tags:")
	log.Printf("   Object: %s/%s", bucketName, objectName)
	log.Println()

	if len(tags) == 0 {
		log.Println("   No tags found for this object")
	} else {
		log.Printf("   Found %d tags:", len(tags))
		for key, value := range tags {
			log.Printf("     %s: %s", key, value)
		}
	}
}
