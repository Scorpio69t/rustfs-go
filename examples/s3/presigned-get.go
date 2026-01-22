//go:build example
// +build example

// Example: Generate a presigned GET URL
// Demonstrates how to generate a presigned URL for temporary access to an object
package main

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go"
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

	// 获取 Object 服务
	objectSvc := client.Object()

	// 要生成预签名 URL 的对象
	bucketName := YOURBUCKET
	objectName := "my-test-object.txt"

	// Set URL expiration (7 days)
	expires := 7 * 24 * time.Hour

	// Optional: add request parameters (e.g. response header overrides)
	reqParams := make(url.Values)
	// Example header overrides:
	// reqParams.Set("response-content-type", "application/json")
	// reqParams.Set("response-content-disposition", "attachment; filename=\"downloaded-file.txt\"")

	// 生成预签名 GET URL
	presignedURL, headers, err := objectSvc.PresignGet(ctx, bucketName, objectName, expires, reqParams)
	if err != nil {
		log.Fatalf("Failed to generate presigned URL: %v", err)
	}

	// Show results
	log.Println("✅ Presigned URL generated successfully!")
	log.Printf("   Object: %s/%s", bucketName, objectName)
	log.Printf("   Expires in: %v", expires)
	log.Println("\nPresigned URL:")
	log.Println("----------------------------------------")
	log.Println(presignedURL.String())
	log.Println("----------------------------------------")

	// Show signed request headers (if any)
	if len(headers) > 0 {
		log.Println("\nSigned request headers:")
		for key, values := range headers {
			for _, value := range values {
				log.Printf("   %s: %s", key, value)
			}
		}
	}

	log.Println("\nHow to use:")
	log.Println("1. Copy the URL above")
	log.Println("2. Open it in a browser or use curl to access:")
	log.Printf("   curl -X GET \"%s\"", presignedURL.String())
	log.Println("\nNote: This URL will expire after 7 days")
}
