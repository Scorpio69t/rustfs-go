//go:build example
// +build example

// 示例：获取对象信息
// 演示如何使用 RustFS Go SDK 获取对象的元数据信息
package main

import (
	"context"
	"log"

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

	// 要查询的对象
	bucketName := YOURBUCKET
	objectName := "my-test-object.txt"

	// 获取对象信息
	objInfo, err := objectSvc.Stat(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("获取对象信息失败: %v", err)
	}

	// 显示对象信息
	log.Println("✅ 对象信息:")
	log.Println("----------------------------------------")
	log.Printf("  对象名: %s", objInfo.Key)
	log.Printf("  存储桶: %s", bucketName)
	log.Printf("  大小: %d 字节", objInfo.Size)
	log.Printf("  内容类型: %s", objInfo.ContentType)
	log.Printf("  ETag: %s", objInfo.ETag)
	log.Printf("  最后修改: %s", objInfo.LastModified.Format("2006-01-02 15:04:05"))

	if objInfo.VersionID != "" {
		log.Printf("  版本ID: %s", objInfo.VersionID)
	}

	// 显示用户元数据
	if len(objInfo.UserMetadata) > 0 {
		log.Println("\n  用户元数据:")
		for key, value := range objInfo.UserMetadata {
			log.Printf("    %s: %s", key, value)
		}
	}

	// 显示标签
	if len(objInfo.UserTags) > 0 {
		log.Println("\n  对象标签:")
		for key, value := range objInfo.UserTags {
			log.Printf("    %s: %s", key, value)
		}
	}

	log.Println("----------------------------------------")
}
