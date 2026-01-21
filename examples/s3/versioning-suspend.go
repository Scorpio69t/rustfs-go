//go:build example
// +build example

// Example: Suspend bucket versioning
// Demonstrates how to use the RustFS Go SDK to suspend a bucket's versioning
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
		log.Fatalf("Unable to create client: %v", err)
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
		log.Fatalf("Failed to suspend versioning: %v", err)
	}

	log.Printf("✅ Bucket '%s' versioning suspended", bucketName)

	// Verify versioning status
	config, err := bucketSvc.GetVersioning(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to get versioning status: %v", err)
	}

	log.Println("\nCurrent versioning status:")
	log.Printf("  Status: %s", config.Status)
	if config.IsSuspended() {
		log.Println("  ⏸️  Versioning is suspended")
	}
}
