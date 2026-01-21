//go:build example
// +build example

// 示例：列出所有存储桶
// 演示如何使用 RustFS Go SDK 列出账户下的所有存储桶
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

	// 列出所有存储桶
	buckets, err := bucketSvc.List(ctx)
	if err != nil {
		log.Fatalf("列出存储桶失败: %v", err)
	}

	// 显示结果
	log.Printf("找到 %d 个存储桶:", len(buckets))
	log.Println("----------------------------------------")

	for i, bucket := range buckets {
		log.Printf("%d. 名称: %s", i+1, bucket.Name)
		log.Printf("   创建时间: %s", bucket.CreationDate.Format("2006-01-02 15:04:05"))
		if bucket.Region != "" {
			log.Printf("   区域: %s", bucket.Region)
		}
		log.Println("----------------------------------------")
	}

	if len(buckets) == 0 {
		log.Println("当前没有存储桶")
	}
}
