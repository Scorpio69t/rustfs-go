//go:build example
// +build example

// Example: Enable bucket versioning
// Demonstrates how to use the RustFS Go SDK to enable bucket versioning
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
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
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

	// 设置版本控制配置
	versioningConfig := types.VersioningConfig{
		Status: "Enabled",
	}

	err = bucketSvc.SetVersioning(ctx, bucketName, versioningConfig)
	if err != nil {
		log.Fatalf("Failed to enable versioning: %v", err)
	}

	log.Printf("✅ Bucket '%s' versioning enabled", bucketName)

	// Verify versioning status
	config, err := bucketSvc.GetVersioning(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to get versioning status: %v", err)
	}

	log.Println("\nCurrent versioning status:")
	log.Printf("  Status: %s", config.Status)
	if config.IsEnabled() {
		log.Println("  ✅ Versioning is enabled")
	}
}
