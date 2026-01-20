//go:build example
// +build example

// 示例：删除存储桶策略
// 演示如何使用 RustFS Go SDK 删除存储桶的访问策略
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

	bucketName := YOURBUCKET

	// 删除存储桶策略
	err = bucketSvc.DeletePolicy(ctx, bucketName)
	if err != nil {
		log.Fatalf("删除存储桶策略失败: %v", err)
	}

	log.Printf("✅ 存储桶 '%s' 的策略已删除", bucketName)
	log.Println("\n提示：存储桶现在没有公共访问策略")
}
