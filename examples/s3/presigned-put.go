//go:build example
// +build example

// Example: Generate a presigned PUT URL
// Demonstrates how to generate a presigned URL for temporary uploads
package main

import (
	"context"
	"log"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// 配置连接参数
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

	// 要生成预签名 URL 的对象
	bucketName := YOURBUCKET
	objectName := "presigned-upload.txt"

	// Set URL expiration (1 hour)
	expires := 1 * time.Hour

	// 生成预签名 PUT URL
	presignedURL, headers, err := objectSvc.PresignPut(ctx, bucketName, objectName, expires, nil)
	if err != nil {
		log.Fatalf("Failed to generate presigned URL: %v", err)
	}

	// Show results
	log.Println("✅ Presigned PUT URL generated successfully!")
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
	log.Println("1. Upload a file with curl:")
	log.Printf("   curl -X PUT -T <local-file> \"%s\"", presignedURL.String())
	log.Println("\n2. Or send a PUT request with an HTTP client")
	log.Println("\nNote: This URL will expire after 1 hour")
}
