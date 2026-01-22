//go:build example
// +build example

// Example: Delete object tagging
// Demonstrates how to delete all tags for an object using the RustFS Go SDK
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

	// 删除对象标签
	err = objectSvc.DeleteTagging(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to delete object tagging: %v", err)
	}

	log.Println("✅ Object tags deleted successfully!")
	log.Printf("   Object: %s/%s", bucketName, objectName)
	log.Println("\nNote: All tags have been removed")
}
