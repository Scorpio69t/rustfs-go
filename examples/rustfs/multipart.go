//go:build example
// +build example

// multipart-new.go - 使用新 API 的分片上传示例
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/types"
)

func main() {
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "mybucket"
	)

	// 初始化客户端
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln("初始化客户端失败:", err)
	}

	ctx := context.Background()
	bucketName := YOURBUCKET
	objectName := "large-file.txt"

	// ===== 使用新 API 进行分片上传 =====
	// 获取 Object 服务
	objectSvc := client.Object()

	// 类型断言以访问分片上传方法
	type MultipartService interface {
		InitiateMultipartUpload(ctx context.Context, bucketName, objectName string, opts ...object.PutOption) (string, error)
		UploadPart(ctx context.Context, bucketName, objectName, uploadID string, partNumber int, reader io.Reader, partSize int64, opts ...object.PutOption) (types.ObjectPart, error)
		CompleteMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string, parts []types.ObjectPart, opts ...object.PutOption) (types.UploadInfo, error)
		AbortMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string) error
	}

	multipartSvc, ok := objectSvc.(MultipartService)
	if !ok {
		log.Fatalln("对象服务不支持分片上传")
	}

	// 1. 初始化分片上传
	log.Println("\n=== 初始化分片上传 ===")
	uploadID, err := multipartSvc.InitiateMultipartUpload(ctx, bucketName, objectName,
		object.WithContentType("text/plain"),
		object.WithUserMetadata(map[string]string{
			"upload-type": "multipart",
		}),
	)
	if err != nil {
		log.Fatalln("初始化分片上传失败:", err)
	}
	log.Printf("✅ 初始化成功，Upload ID: %s\n", uploadID)

	// 延迟取消（如果出错）
	var uploadCompleted bool
	defer func() {
		if !uploadCompleted {
			log.Println("\n=== 取消分片上传（清理）===")
			err := multipartSvc.AbortMultipartUpload(ctx, bucketName, objectName, uploadID)
			if err != nil {
				log.Printf("取消分片上传失败: %v\n", err)
			} else {
				log.Println("✅ 已取消分片上传")
			}
		}
	}()

	// 2. 上传分片
	log.Println("\n=== 上传分片 ===")
	parts := make([]types.ObjectPart, 0)

	// 模拟 3 个分片（每个分片至少 5MB，最后一个可以小于 5MB）
	// 注意：S3 要求每个分片（除了最后一个）至少 5MB
	partContents := []string{
		strings.Repeat("Part 1: This is the first part of the file. ", 120000),          // ~5.3MB
		strings.Repeat("Part 2: This is the second part of the file. ", 120000),         // ~5.4MB
		strings.Repeat("Part 3: This is the third and final part of the file. ", 50000), // ~2.5MB (最后一个可以小于5MB)
	}

	for i, content := range partContents {
		partNumber := i + 1
		partData := strings.NewReader(content)
		partSize := int64(len(content))

		log.Printf("上传分片 %d/%d (大小: %d bytes)...\n", partNumber, len(partContents), partSize)

		part, err := multipartSvc.UploadPart(ctx, bucketName, objectName, uploadID,
			partNumber, partData, partSize)
		if err != nil {
			log.Fatalf("上传分片 %d 失败: %v\n", partNumber, err)
		}

		parts = append(parts, part)
		log.Printf("  ✅ 分片 %d 上传成功，ETag: %s\n", partNumber, part.ETag)
	}

	// 3. 完成分片上传
	log.Println("\n=== 完成分片上传 ===")
	uploadInfo, err := multipartSvc.CompleteMultipartUpload(ctx, bucketName, objectName, uploadID, parts)
	if err != nil {
		log.Fatalln("完成分片上传失败:", err)
	}

	uploadCompleted = true // 标记上传已完成，避免取消
	log.Printf("✅ 分片上传完成！\n")
	log.Printf("   对象: %s\n", uploadInfo.Key)
	log.Printf("   ETag: %s\n", uploadInfo.ETag)
	log.Printf("   总大小: %d bytes\n", uploadInfo.Size)

	// 4. 验证上传的对象
	log.Println("\n=== 验证上传的对象 ===")
	objInfo, err := objectSvc.Stat(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalln("获取对象信息失败:", err)
	}
	log.Printf("对象信息:\n")
	log.Printf("  名称: %s\n", objInfo.Key)
	log.Printf("  大小: %d bytes\n", objInfo.Size)
	log.Printf("  ETag: %s\n", objInfo.ETag)
	log.Printf("  修改时间: %s\n", objInfo.LastModified.Format("2006-01-02 15:04:05"))

	// 5. 下载并显示部分内容
	log.Println("\n=== 下载并显示部分内容 ===")
	reader, _, err := objectSvc.Get(ctx, bucketName, objectName,
		object.WithGetRange(0, 99), // 下载前 100 字节
	)
	if err != nil {
		log.Fatalln("下载对象失败:", err)
	}
	defer reader.Close()

	buf := make([]byte, 100)
	n, _ := reader.Read(buf)
	log.Printf("前 100 字节内容:\n%s\n", string(buf[:n]))

	// 6. 清理（可选）
	// log.Println("\n=== 删除上传的对象 ===")
	// err = objectSvc.Delete(ctx, bucketName, objectName)
	// if err != nil {
	// 	log.Printf("删除对象失败: %v\n", err)
	// } else {
	// 	log.Printf("✅ 成功删除对象: %s\n", objectName)
	// }

	log.Println("\n=== 分片上传示例运行完成 ===")
	fmt.Println("\n提示：")
	fmt.Println("- 分片上传适用于大文件（>5MB）")
	fmt.Println("- 每个分片最小 5MB（最后一个分片除外）")
	fmt.Println("- 最多支持 10,000 个分片")
	fmt.Println("- 如果上传失败，已上传的分片会自动清理")
}
