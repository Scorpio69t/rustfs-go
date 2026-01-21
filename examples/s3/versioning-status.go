//go:build example
// +build example

// Example: Get bucket versioning status
// Demonstrates how to use the RustFS Go SDK to get a bucket's versioning status
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
		log.Fatalf("Unable to create client: %v", err)
	}

	ctx := context.Background()

	// 获取 Bucket 服务
	bucketSvc := client.Bucket()

	bucketName := YOURBUCKET

	// 获取版本控制状态
	config, err := bucketSvc.GetVersioning(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to get versioning status: %v", err)
	}

	log.Printf("✅ Bucket '%s' versioning status:", bucketName)
	log.Println("----------------------------------------")
	log.Printf("  Status: %s", config.Status)

	if config.MFADelete != "" {
		log.Printf("  MFA Delete: %s", config.MFADelete)
	}

	log.Println()
	if config.IsEnabled() {
		log.Println("  ✅ Versioning is enabled")
		log.Println("  New object uploads will receive a version ID")
	} else if config.IsSuspended() {
		log.Println("  ⏸️  Versioning is suspended")
		log.Println("  New object uploads will not receive a version ID")
	} else {
		log.Println("  ❌ Versioning is not enabled")
	}
	log.Println("----------------------------------------")
}
