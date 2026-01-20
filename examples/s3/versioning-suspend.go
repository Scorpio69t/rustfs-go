//go:build example
// +build example

// 示例：暂停存储桶版本控制
// 演示如何使用 RustFS Go SDK 暂停存储桶的版本控制
package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/types"
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

	// 设置版本控制配置为暂停
	versioningConfig := types.VersioningConfig{
		Status: "Suspended",
	}

	err = bucketSvc.SetVersioning(ctx, bucketName, versioningConfig)
	if err != nil {
		log.Fatalf("暂停版本控制失败: %v", err)
	}

	log.Printf("✅ 存储桶 '%s' 的版本控制已暂停", bucketName)

	// 验证版本控制状态
	config, err := bucketSvc.GetVersioning(ctx, bucketName)
	if err != nil {
		log.Fatalf("获取版本控制状态失败: %v", err)
	}

	log.Println("\n当前版本控制状态:")
	log.Printf("  状态: %s", config.Status)
	if config.IsSuspended() {
		log.Println("  ⏸️  版本控制已暂停")
	}
}
