//go:build example
// +build example

// 示例：删除对象标签
// 演示如何使用 RustFS Go SDK 删除对象的所有标签
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

	// 删除对象标签
	err = objectSvc.DeleteTagging(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("删除对象标签失败: %v", err)
	}

	log.Println("✅ 对象标签删除成功!")
	log.Printf("   对象: %s/%s", bucketName, objectName)
	log.Println("\n提示：所有标签已被删除")
}
