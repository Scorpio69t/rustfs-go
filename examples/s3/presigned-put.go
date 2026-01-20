//go:build example
// +build example

// 示例：生成预签名 PUT URL
// 演示如何使用 RustFS Go SDK 生成预签名 URL 以便临时上传对象
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
		log.Fatalf("无法创建客户端: %v", err)
	}

	ctx := context.Background()

	// 获取 Object 服务
	objectSvc := client.Object()

	// 要生成预签名 URL 的对象
	bucketName := YOURBUCKET
	objectName := "presigned-upload.txt"

	// 设置 URL 过期时间（1小时）
	expires := 1 * time.Hour

	// 生成预签名 PUT URL
	presignedURL, headers, err := objectSvc.PresignPut(ctx, bucketName, objectName, expires, nil)
	if err != nil {
		log.Fatalf("生成预签名 URL 失败: %v", err)
	}

	// 显示结果
	log.Println("✅ 预签名 PUT URL 生成成功!")
	log.Printf("   对象: %s/%s", bucketName, objectName)
	log.Printf("   有效期: %v", expires)
	log.Println("\n预签名 URL:")
	log.Println("----------------------------------------")
	log.Println(presignedURL.String())
	log.Println("----------------------------------------")

	// 显示签名的请求头（如果有）
	if len(headers) > 0 {
		log.Println("\n签名的请求头:")
		for key, values := range headers {
			for _, value := range values {
				log.Printf("   %s: %s", key, value)
			}
		}
	}

	log.Println("\n使用方法:")
	log.Println("1. 使用 curl 上传文件:")
	log.Printf("   curl -X PUT -T <local-file> \"%s\"", presignedURL.String())
	log.Println("\n2. 或使用 HTTP 客户端发送 PUT 请求")
	log.Println("\n注意：此 URL 将在 1 小时后失效")
}
