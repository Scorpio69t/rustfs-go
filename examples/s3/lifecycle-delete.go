//go:build example
// +build example

// 示例：删除存储桶生命周期策略
// 演示如何移除生命周期配置
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

const (
	endpoint  = "127.0.0.1:9000"
	accessKey = "XhJOoEKn3BM6cjD2dVmx"
	secretKey = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
	bucket    = "mybucket"
)

func main() {
	// 创建客户端
	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	service := client.Bucket()

	fmt.Printf("删除存储桶 '%s' 的生命周期配置...\n", bucket)

	// 删除生命周期配置
	err = service.DeleteLifecycle(ctx, bucket)
	if err != nil {
		log.Fatalf("删除生命周期配置失败: %v\n", err)
	}

	fmt.Println("✅ 生命周期配置已删除")
}
