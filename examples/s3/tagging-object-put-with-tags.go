//go:build example
// +build example

// Example: Upload an object with tags
// Demonstrates how to set tags when uploading an object
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

const (
	endpoint  = "127.0.0.1:9000"
	accessKey = "rustfsadmin"
	secretKey = "rustfsadmin"
	bucket    = "mybucket"
)

func main() {
	// Create client
	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	service := client.Object()

	objectName := "tagged-object.txt"
	content := "This is an object with tags"

	// 定义对象标签
	tags := map[string]string{
		"Environment": "Development",
		"Project":     "RustFS-Go",
		"Owner":       "DevTeam",
		"Category":    "Sample",
	}

	fmt.Printf("Uploading object '%s' with tags...\n", objectName)

	// 上传对象时设置标签
	reader := strings.NewReader(content)
	uploadInfo, err := service.Put(
		ctx,
		bucket,
		objectName,
		reader,
		int64(len(content)),
		object.WithContentType("text/plain; charset=utf-8"),
		object.WithUserTags(tags),
	)
	if err != nil {
		log.Fatalf("Failed to upload object: %v\n", err)
	}

	fmt.Printf("✅ Object uploaded successfully\n")
	fmt.Printf("   Key: %s\n", uploadInfo.Key)
	fmt.Printf("   ETag: %s\n", uploadInfo.ETag)
	fmt.Printf("   Size: %d bytes\n", uploadInfo.Size)

	// 读取对象标签以验证
	fmt.Println("\nVerifying object tags...")
	objectTags, err := service.GetTagging(ctx, bucket, objectName)
	if err != nil {
		log.Fatalf("Failed to get tags: %v", err)
	}

	if len(objectTags) == 0 {
		fmt.Println("⚠️  No tags found for object")
		return
	}

	fmt.Println("✅ Object tags:")
	for key, value := range objectTags {
		fmt.Printf("   %s = %s\n", key, value)
	}
}
