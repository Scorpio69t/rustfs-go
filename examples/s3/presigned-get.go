//go:build example
// +build example

// 示例：生成预签名 GET URL
// 演示如何使用 RustFS Go SDK 生成预签名 URL 以便临时访问对象
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
	objectName := "my-test-object.txt"

	// 设置 URL 过期时间（7天）
	expires := 7 * 24 * time.Hour

	// 可选：添加请求参数（如响应头覆盖）
	reqParams := make(url.Values)
	// 覆盖响应头示例：
	// reqParams.Set("response-content-type", "application/json")
	// reqParams.Set("response-content-disposition", "attachment; filename=\"downloaded-file.txt\"")

	// 生成预签名 GET URL
	presignedURL, headers, err := objectSvc.PresignGet(ctx, bucketName, objectName, expires, reqParams)
	if err != nil {
		log.Fatalf("生成预签名 URL 失败: %v", err)
	}

	// 显示结果
	log.Println("✅ 预签名 URL 生成成功!")
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
	log.Println("1. 复制上面的 URL")
	log.Println("2. 在浏览器中打开或使用 curl 访问:")
	log.Printf("   curl -X GET \"%s\"", presignedURL.String())
	log.Println("\n注意：此 URL 将在 7 天后失效")
}
