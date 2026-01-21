//go:build example
// +build example

// 示例：获取存储桶位置
// 演示如何使用 RustFS Go SDK 获取存储桶的区域位置
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

	// 获取存储桶位置
	bucketName := YOURBUCKET
	location, err := bucketSvc.GetLocation(ctx, bucketName)
	if err != nil {
		log.Fatalf("获取存储桶位置失败: %v", err)
	}

	log.Printf("✅ 存储桶 '%s' 位于区域: %s", bucketName, location)

	if location == "" {
		log.Println("提示：空字符串通常表示默认区域（us-east-1）")
	}
}
