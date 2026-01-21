//go:build example
// +build example

// 示例：生成带响应头覆盖的预签名 URL
// 演示如何通过预签名 URL 自定义响应头
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
	accessKey = "XhJOoEKn3BM6cjD2dVmx"
	secretKey = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
	bucket    = "mybucket"
)

func main() {
	// 创建客户端
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

	fmt.Printf("为对象 '%s' 生成带响应头覆盖的预签名 URL...\n", objectName)

	// 生成预签名 GET URL（有效期 15 分钟）
	presignedURL, headers, err := service.PresignGet(
		ctx,
		bucket,
		objectName,
		15*time.Minute,
		params,
	)
	if err != nil {
		log.Fatalf("生成预签名 URL 失败: %v\n", err)
	}

	fmt.Println("\n✅ 预签名 URL 生成成功")
	fmt.Printf("URL: %s\n", presignedURL.String())

	if len(headers) > 0 {
		fmt.Println("\n需要包含的请求头:")
		for key, values := range headers {
			for _, value := range values {
				fmt.Printf("  %s: %s\n", key, value)
			}
		}
	}

	fmt.Println("\n使用此 URL 下载文件时，响应头将被覆盖:")
	fmt.Println("  Content-Type: application/octet-stream")
	fmt.Println("  Content-Disposition: attachment; filename=\"downloaded-file.txt\"")
	fmt.Println("  Cache-Control: no-cache, no-store, must-revalidate")
	fmt.Printf("  Expires: %s\n", params.Get("response-expires"))

	fmt.Println("\n示例用法:")
	fmt.Printf("  curl -O \"%s\"\n", presignedURL.String())
}
