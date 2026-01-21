//go:build example
// +build example

// 示例：设置存储桶生命周期策略
// 演示如何配置对象自动过期和转换规则
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

	// 定义生命周期配置（XML 格式）
	// 规则：temp/ 目录下的对象 30 天后自动删除
	lifecycleConfig := `<LifecycleConfiguration>
	<Rule>
		<ID>expire-temp-files</ID>
		<Status>Enabled</Status>
		<Filter>
			<Prefix>temp/</Prefix>
		</Filter>
		<Expiration>
			<Days>30</Days>
		</Expiration>
	</Rule>
	<Rule>
		<ID>expire-old-logs</ID>
		<Status>Enabled</Status>
		<Filter>
			<Prefix>logs/</Prefix>
		</Filter>
		<Expiration>
			<Days>90</Days>
		</Expiration>
	</Rule>
</LifecycleConfiguration>`

	fmt.Printf("为存储桶 '%s' 设置生命周期策略...\n", bucket)

	// 设置生命周期配置
	err = service.SetLifecycle(ctx, bucket, []byte(lifecycleConfig))
	if err != nil {
		log.Fatalf("设置生命周期配置失败: %v\n", err)
	}

	fmt.Println("✅ 生命周期配置设置成功")
	fmt.Println("\n生命周期规则:")
	fmt.Println("  1. temp/ 目录：30 天后自动删除")
	fmt.Println("  2. logs/ 目录：90 天后自动删除")
}
