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

	// 定义要删除的对象列表
	objectsToDelete := []string{
		"test-object-1.txt",
		"test-object-2.txt",
		"test-object-3.txt",
	}

	fmt.Printf("准备批量删除 %d 个对象...\n", len(objectsToDelete))

	// 逐个删除对象（当前 API 暂不支持批量删除，这里演示循环删除）
	deletedCount := 0
	for _, objectName := range objectsToDelete {
		err := service.Delete(ctx, bucket, objectName)
		if err != nil {
			fmt.Printf("⚠️  删除对象 '%s' 失败: %v\n", objectName, err)
			continue
		}
		deletedCount++
		fmt.Printf("✅ 已删除对象: %s\n", objectName)
	}

	fmt.Printf("\n删除完成：成功 %d 个，失败 %d 个\n", deletedCount, len(objectsToDelete)-deletedCount)
}
