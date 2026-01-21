//go:build example
// +build example

// 示例：复制对象
// 演示如何使用 RustFS Go SDK 复制对象
package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
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

	// 获取 Object 服务
	objectSvc := client.Object()

	// 源对象
	srcBucket := YOURBUCKET
	srcObject := "my-test-object.txt"

	// 目标对象
	destBucket := YOURBUCKET
	destObject := "my-test-object-copy.txt"

	// 复制对象
	copyInfo, err := objectSvc.Copy(ctx,
		destBucket, destObject, // 目标
		srcBucket, srcObject, // 源
	)
	if err != nil {
		log.Fatalf("复制对象失败: %v", err)
	}

	log.Println("✅ 对象复制成功!")
	log.Printf("   源: %s/%s", srcBucket, srcObject)
	log.Printf("   目标: %s/%s", destBucket, destObject)
	log.Printf("   ETag: %s", copyInfo.ETag)

	if copyInfo.VersionID != "" {
		log.Printf("   版本ID: %s", copyInfo.VersionID)
	}

	// 示例：复制时添加新元数据
	log.Println("\n=== 复制并替换元数据 ===")
	destObject2 := "my-test-object-copy-with-metadata.txt"

	copyInfo2, err := objectSvc.Copy(ctx,
		destBucket, destObject2,
		srcBucket, srcObject,
		object.WithCopyMetadata(map[string]string{
			"copied-at": "2026-01-20",
			"author":    "rustfs-sdk",
		}, true), // true 表示替换元数据
	)
	if err != nil {
		log.Printf("复制对象失败: %v", err)
	} else {
		log.Printf("✅ 复制成功: %s", destObject2)
		log.Printf("   ETag: %s", copyInfo2.ETag)
	}
}
