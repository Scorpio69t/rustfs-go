//go:build example
// +build example

// 示例：下载对象
// 演示如何使用 RustFS Go SDK 下载对象
package main

import (
	"context"
	"io"
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

	// 要下载的对象
	bucketName := YOURBUCKET
	objectName := "my-test-object.txt"

	// 下载对象
	reader, objInfo, err := objectSvc.Get(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("下载对象失败: %v", err)
	}
	defer reader.Close()

	// 显示对象信息
	log.Println("✅ 对象下载成功!")
	log.Printf("   对象名: %s", objInfo.Key)
	log.Printf("   大小: %d 字节", objInfo.Size)
	log.Printf("   类型: %s", objInfo.ContentType)
	log.Printf("   ETag: %s", objInfo.ETag)
	log.Printf("   修改时间: %s", objInfo.LastModified.Format("2006-01-02 15:04:05"))

	// 显示用户元数据（如果有）
	if len(objInfo.UserMetadata) > 0 {
		log.Println("   用户元数据:")
		for key, value := range objInfo.UserMetadata {
			log.Printf("     %s: %s", key, value)
		}
	}

	// 读取内容并显示
	log.Println("\n对象内容:")
	log.Println("----------------------------------------")

	content, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("读取对象内容失败: %v", err)
	}

	log.Println(string(content))
	log.Println("----------------------------------------")
}
