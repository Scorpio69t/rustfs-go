//go:build example
// +build example

// 示例：删除对象
// 演示如何使用 RustFS Go SDK 删除一个对象
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

	// 要删除的对象
	bucketName := YOURBUCKET
	objectName := "object-to-delete.txt"

	// 删除对象
	err = objectSvc.Delete(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("删除对象失败: %v", err)
	}

	log.Printf("✅ 成功删除对象: %s/%s", bucketName, objectName)
	log.Println("\n注意：删除不存在的对象不会报错")
}
