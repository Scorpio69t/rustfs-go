//go:build example
// +build example

// 示例：删除存储桶
// 演示如何使用 RustFS Go SDK 删除一个存储桶
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

	// 要删除的存储桶名称
	bucketName := "test-bucket-to-delete"

	// 删除存储桶
	// 注意：存储桶必须为空才能删除
	err = bucketSvc.Delete(ctx, bucketName)
	if err != nil {
		log.Fatalf("删除存储桶失败: %v", err)
	}

	log.Printf("✅ 成功删除存储桶: %s", bucketName)
	log.Println("\n注意：只有空存储桶才能被删除")
	log.Println("如果存储桶包含对象，请先删除所有对象")
}
