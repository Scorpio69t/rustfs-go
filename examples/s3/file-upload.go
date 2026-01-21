//go:build example
// +build example

// 示例：从文件上传对象
// 演示如何使用 RustFS Go SDK 将本地文件上传为对象
package main

import (
	"context"
	"log"
	"os"

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

	// 上传参数
	bucketName := YOURBUCKET
	objectName := "uploaded-file.txt"
	filePath := "test-file.txt" // 要上传的本地文件路径

	// 创建测试文件（如果不存在）
	if err := createTestFile(filePath); err != nil {
		log.Fatalf("创建测试文件失败: %v", err)
	}

	// 从文件上传对象
	// FPut 会自动检测文件大小和内容类型
	uploadInfo, err := objectSvc.FPut(ctx, bucketName, objectName, filePath,
		object.WithContentType("text/plain"),
		object.WithUserMetadata(map[string]string{
			"source": "local-file",
		}),
	)
	if err != nil {
		log.Fatalf("文件上传失败: %v", err)
	}

	// 显示上传结果
	log.Println("✅ 文件上传成功!")
	log.Printf("   本地文件: %s", filePath)
	log.Printf("   存储桶: %s", uploadInfo.Bucket)
	log.Printf("   对象名: %s", uploadInfo.Key)
	log.Printf("   ETag: %s", uploadInfo.ETag)
	log.Printf("   大小: %d 字节", uploadInfo.Size)

	if uploadInfo.VersionID != "" {
		log.Printf("   版本ID: %s", uploadInfo.VersionID)
	}

	log.Println("\n提示：使用 file-download.go 示例下载此对象到文件")
}

// createTestFile 创建测试文件
func createTestFile(filePath string) error {
	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		log.Printf("使用现有文件: %s", filePath)
		return nil
	}

	// 创建新文件
	content := "This is a test file for upload demonstration.\n" +
		"RustFS Go SDK - File Upload Example\n" +
		"Generated automatically for testing purposes.\n"

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return err
	}

	log.Printf("已创建测试文件: %s", filePath)
	return nil
}
