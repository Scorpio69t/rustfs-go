//go:build example
// +build example

// 示例：列出对象（V2 API）
// 演示如何使用 RustFS Go SDK 列出存储桶中的对象
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

	bucketName := YOURBUCKET

	// 列出所有对象
	log.Printf("=== 列出存储桶 '%s' 中的所有对象 ===\n", bucketName)
	objectsCh := objectSvc.List(ctx, bucketName)

	count := 0
	for obj := range objectsCh {
		if obj.Err != nil {
			log.Fatalf("列出对象失败: %v", obj.Err)
		}
		count++

		if obj.IsPrefix {
			log.Printf("%d. %s (目录)", count, obj.Key)
		} else {
			log.Printf("%d. %s", count, obj.Key)
			log.Printf("   大小: %d 字节", obj.Size)
			log.Printf("   修改时间: %s", obj.LastModified.Format("2006-01-02 15:04:05"))
			log.Printf("   ETag: %s", obj.ETag)
			log.Println()
		}
	}

	if count == 0 {
		log.Println("存储桶为空")
	} else {
		log.Printf("\n总共找到 %d 个对象", count)
	}

	// 示例：使用前缀过滤
	log.Println("\n=== 列出前缀为 'my-test' 的对象 ===")
	objectsCh = objectSvc.List(ctx, bucketName,
		object.WithListPrefix("my-test"),
	)

	count = 0
	for obj := range objectsCh {
		if obj.Err != nil {
			log.Printf("列出对象失败: %v", obj.Err)
			break
		}
		count++
		log.Printf("%d. %s (%d 字节)", count, obj.Key, obj.Size)
	}

	if count == 0 {
		log.Println("没有找到匹配的对象")
	}
}
