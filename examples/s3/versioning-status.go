//go:build example
// +build example

// 示例：获取存储桶版本控制状态
// 演示如何使用 RustFS Go SDK 获取存储桶的版本控制状态
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

	// 获取版本控制状态
	config, err := bucketSvc.GetVersioning(ctx, bucketName)
	if err != nil {
		log.Fatalf("获取版本控制状态失败: %v", err)
	}

	log.Printf("✅ 存储桶 '%s' 版本控制状态:", bucketName)
	log.Println("----------------------------------------")
	log.Printf("  状态: %s", config.Status)

	if config.MFADelete != "" {
		log.Printf("  MFA 删除: %s", config.MFADelete)
	}

	log.Println()
	if config.IsEnabled() {
		log.Println("  ✅ 版本控制已启用")
		log.Println("  新上传的对象将生成版本ID")
	} else if config.IsSuspended() {
		log.Println("  ⏸️  版本控制已暂停")
		log.Println("  新上传的对象不会生成版本ID")
	} else {
		log.Println("  ❌ 版本控制未启用")
	}
	log.Println("----------------------------------------")
}
