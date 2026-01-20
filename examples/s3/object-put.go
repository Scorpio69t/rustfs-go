//go:build example
// +build example

// 示例：上传对象
// 演示如何使用 RustFS Go SDK 上传对象到存储桶
package main

import (
	"context"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
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

	// 准备上传数据
	bucketName := YOURBUCKET
	objectName := "my-test-object.txt"
	content := "Hello, RustFS! This is a test object uploaded using the RustFS Go SDK."

	// 创建 Reader
	reader := strings.NewReader(content)

	// 上传对象
	// 使用选项函数设置内容类型和用户元数据
	uploadInfo, err := objectSvc.Put(ctx, bucketName, objectName, reader, int64(len(content)),
		object.WithContentType("text/plain; charset=utf-8"),
		object.WithUserMetadata(map[string]string{
			"author":      "rustfs-go-sdk",
			"description": "示例对象",
		}),
	)
	if err != nil {
		log.Fatalf("上传对象失败: %v", err)
	}

	// 显示上传结果
	log.Println("✅ 对象上传成功!")
	log.Printf("   存储桶: %s", uploadInfo.Bucket)
	log.Printf("   对象名: %s", uploadInfo.Key)
	log.Printf("   ETag: %s", uploadInfo.ETag)
	log.Printf("   大小: %d 字节", uploadInfo.Size)

	if uploadInfo.VersionID != "" {
		log.Printf("   版本ID: %s", uploadInfo.VersionID)
	}

	log.Println("\n提示：使用 object-get.go 示例下载此对象")
}
