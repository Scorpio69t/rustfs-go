//go:build example
// +build example

// Example: Set object tagging
// Demonstrates how to set tags for an object using the RustFS Go SDK
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
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
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

	// Define tags
	tags := map[string]string{
		"environment": "production",
		"project":     "rustfs-sdk",
		"owner":       "devteam",
		"version":     "1.0",
	}

	// Set object tags
	err = objectSvc.SetTagging(ctx, bucketName, objectName, tags)
	if err != nil {
		log.Fatalf("Failed to set object tags: %v", err)
	}

	log.Println("✅ Object tags set successfully!")
	log.Printf("   Object: %s/%s", bucketName, objectName)
	log.Println("\nTags set:")
	for key, value := range tags {
		log.Printf("   %s: %s", key, value)
	}

	log.Println("\nTip: use tagging-object-get.go to view tags")
}
