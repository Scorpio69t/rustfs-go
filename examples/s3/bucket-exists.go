//go:build example
// +build example

// 示例：检查存储桶是否存在
// 演示如何使用 RustFS Go SDK 检查存储桶是否存在
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

	// 获取 Bucket 服务
	bucketSvc := client.Bucket()

	// 检查存储桶是否存在
	bucketName := YOURBUCKET
	exists, err := bucketSvc.Exists(ctx, bucketName)
	if err != nil {
		log.Fatalf("检查存储桶失败: %v", err)
	}

	if exists {
		log.Printf("✅ 存储桶 '%s' 存在", bucketName)
	} else {
		log.Printf("❌ 存储桶 '%s' 不存在", bucketName)
	}

	// 检查一个不存在的存储桶
	nonExistentBucket := "this-bucket-does-not-exist-12345"
	exists, err = bucketSvc.Exists(ctx, nonExistentBucket)
	if err != nil {
		log.Fatalf("检查存储桶失败: %v", err)
	}

	if exists {
		log.Printf("✅ 存储桶 '%s' 存在", nonExistentBucket)
	} else {
		log.Printf("❌ 存储桶 '%s' 不存在", nonExistentBucket)
	}
}
