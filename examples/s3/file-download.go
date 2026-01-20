//go:build example
// +build example

// 示例：下载对象到文件
// 演示如何使用 RustFS Go SDK 将对象下载到本地文件
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

	// 下载参数
	bucketName := YOURBUCKET
	objectName := "my-test-object.txt"
	localFilePath := "downloaded-file.txt"

	// 下载对象到文件
	objInfo, err := objectSvc.FGet(ctx, bucketName, objectName, localFilePath)
	if err != nil {
		log.Fatalf("下载文件失败: %v", err)
	}

	// 显示下载结果
	log.Println("✅ 文件下载成功!")
	log.Printf("   对象名: %s", objInfo.Key)
	log.Printf("   保存到: %s", localFilePath)
	log.Printf("   大小: %d 字节", objInfo.Size)
	log.Printf("   类型: %s", objInfo.ContentType)
	log.Printf("   ETag: %s", objInfo.ETag)
	log.Printf("   修改时间: %s", objInfo.LastModified.Format("2006-01-02 15:04:05"))
}
