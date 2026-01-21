//go:build example
// +build example

// 示例：获取存储桶生命周期策略
// 演示如何查询当前的生命周期配置
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

	fmt.Printf("获取存储桶 '%s' 的生命周期配置...\n\n", bucket)

	// 获取生命周期配置
	config, err := service.GetLifecycle(ctx, bucket)
	if err != nil {
		log.Fatalf("获取生命周期配置失败: %v\n", err)
	}

	if len(config) == 0 {
		fmt.Println("该存储桶没有配置生命周期策略")
		return
	}

	fmt.Println("✅ 生命周期配置:")
	fmt.Println(string(config))
}
