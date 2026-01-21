//go:build example
// +build example

// 示例：获取对象标签
// 演示如何使用 RustFS Go SDK 获取对象的标签
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

	// 获取 Object 服务
	objectSvc := client.Object()

	// 对象信息
	bucketName := YOURBUCKET
	objectName := "my-test-object.txt"

	// 获取对象标签
	tags, err := objectSvc.GetTagging(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("获取对象标签失败: %v", err)
	}

	log.Println("✅ 对象标签:")
	log.Printf("   对象: %s/%s", bucketName, objectName)
	log.Println()

	if len(tags) == 0 {
		log.Println("   该对象没有标签")
	} else {
		log.Printf("   找到 %d 个标签:", len(tags))
		for key, value := range tags {
			log.Printf("     %s: %s", key, value)
		}
	}
}
