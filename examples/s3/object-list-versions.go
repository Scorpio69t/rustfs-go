//go:build example
// +build example

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
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	service := client.Object()

	fmt.Printf("列出存储桶 '%s' 中的所有对象版本...\n\n", bucket)

	// 列出对象版本
	// 使用 ListVersions 方法列出所有版本
	objectCh := service.ListVersions(ctx, bucket)

	versionCount := 0
	currentCount := 0

	for obj := range objectCh {
		if obj.Err != nil {
			fmt.Printf("错误: %v\n", obj.Err)
			continue
		}

		if obj.IsLatest {
			currentCount++
			fmt.Printf("📄 对象: %s\n", obj.Key)
			fmt.Printf("   版本ID: %s (当前版本)\n", obj.VersionID)
		} else {
			versionCount++
			fmt.Printf("📋 对象: %s\n", obj.Key)
			fmt.Printf("   版本ID: %s\n", obj.VersionID)
		}

		fmt.Printf("   大小: %d 字节\n", obj.Size)
		fmt.Printf("   最后修改: %s\n", obj.LastModified.Format("2006-01-02 15:04:05"))
		if obj.IsDeleteMarker {
			fmt.Printf("   ⚠️  删除标记\n")
		}
		fmt.Println()
	}

	fmt.Printf("总计: %d 个当前版本, %d 个历史版本\n", currentCount, versionCount)
}
