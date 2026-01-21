//go:build example
// +build example

// 示例：上传带标签的对象
// 演示如何在上传对象时直接设置标签
package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
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
	service := client.Object()

	objectName := "tagged-object.txt"
	content := "这是一个带标签的对象"

	// 定义对象标签
	tags := map[string]string{
		"Environment": "Development",
		"Project":     "RustFS-Go",
		"Owner":       "DevTeam",
		"Category":    "Sample",
	}

	fmt.Printf("上传对象 '%s' 并设置标签...\n", objectName)

	// 上传对象时设置标签
	reader := strings.NewReader(content)
	uploadInfo, err := service.Put(
		ctx,
		bucket,
		objectName,
		reader,
		int64(len(content)),
		object.WithContentType("text/plain; charset=utf-8"),
		object.WithUserTags(tags),
	)
	if err != nil {
		log.Fatalf("上传对象失败: %v\n", err)
	}

	fmt.Printf("✅ 对象上传成功\n")
	fmt.Printf("   对象名: %s\n", uploadInfo.Key)
	fmt.Printf("   ETag: %s\n", uploadInfo.ETag)
	fmt.Printf("   大小: %d 字节\n", uploadInfo.Size)

	// 读取对象标签以验证
	fmt.Println("\n验证对象标签...")
	objectTags, err := service.GetTagging(ctx, bucket, objectName)
	if err != nil {
		log.Fatalf("获取标签失败: %v\n", err)
	}

	if len(objectTags) == 0 {
		fmt.Println("⚠️  对象没有标签")
		return
	}

	fmt.Println("✅ 对象标签:")
	for key, value := range objectTags {
		fmt.Printf("   %s = %s\n", key, value)
	}
}
