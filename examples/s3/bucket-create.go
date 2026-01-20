//go:build example
// +build example

// 示例：创建存储桶
// 演示如何使用 RustFS Go SDK 创建一个新的存储桶
package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/bucket"
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
		Secure:      false, // 本地测试使用 HTTP，生产环境设置为 true
	})
	if err != nil {
		log.Fatalf("无法创建客户端: %v", err)
	}

	ctx := context.Background()

	// 获取 Bucket 服务
	bucketSvc := client.Bucket()

	// 要创建的存储桶名称
	bucketName := YOURBUCKET

	// 创建存储桶
	// 使用选项函数设置区域
	err = bucketSvc.Create(ctx, bucketName,
		bucket.WithRegion("us-east-1"),
	)
	if err != nil {
		log.Fatalf("创建存储桶失败: %v", err)
	}

	log.Printf("✅ 成功创建存储桶: %s", bucketName)

	// 验证存储桶是否存在
	exists, err := bucketSvc.Exists(ctx, bucketName)
	if err != nil {
		log.Fatalf("检查存储桶失败: %v", err)
	}

	if exists {
		log.Printf("✅ 存储桶 '%s' 已确认存在", bucketName)
	}
}
