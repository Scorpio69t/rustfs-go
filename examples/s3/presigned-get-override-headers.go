//go:build example
// +build example

// Example: Generate a presigned URL with response header overrides
// Demonstrates how to customize response headers via a presigned URL
package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go"
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

	objectName := "my-test-object.txt"

	// 设置响应头覆盖参数
	// 这些参数将改变下载时的响应头
	params := url.Values{}
	params.Set("response-content-type", "application/octet-stream")
	params.Set("response-content-disposition", "attachment; filename=\"downloaded-file.txt\"")
	params.Set("response-cache-control", "no-cache, no-store, must-revalidate")
	params.Set("response-expires", time.Now().Add(1*time.Hour).Format(time.RFC1123))

	fmt.Printf("Generating a presigned URL with response header overrides for object '%s'...\n", objectName)

	// Generate presigned GET URL (expires in 15 minutes)
	presignedURL, headers, err := service.PresignGet(
		ctx,
		bucket,
		objectName,
		15*time.Minute,
		params,
	)
	if err != nil {
		log.Fatalf("Failed to generate presigned URL: %v\n", err)
	}
	fmt.Println("\n✅ Presigned URL generated")
	fmt.Printf("URL: %s\n", presignedURL.String())

	if len(headers) > 0 {
		fmt.Println("\nRequired request headers:")
		for key, values := range headers {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	fmt.Println("\nWhen downloading with this URL, response headers will be overridden:")
	fmt.Println("  Content-Type: application/octet-stream")
	fmt.Println("  Content-Disposition: attachment; filename=\"downloaded-file.txt\"")
	fmt.Println("  Cache-Control: no-cache, no-store, must-revalidate")
	fmt.Printf("  Expires: %s\n", params.Get("response-expires"))

	fmt.Println("\nExample usage:")
	fmt.Printf("  curl -O \"%s\"\n", presignedURL.String())
}
