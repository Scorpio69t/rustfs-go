//go:build example
// +build example

// objectops-new.go - 使用新 API 的对象操作示例
package main

import (
	"context"
	"io"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "mybucket" // 使用 'mc mb play/mybucket' 创建存储桶（如果不存在）
	)

	// 初始化客户端
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false, // 设置为 true 使用 HTTPS
	})
	if err != nil {
		log.Fatalln("初始化客户端失败:", err)
	}

	ctx := context.Background()
	bucketName := YOURBUCKET
	objectName := "test-object.txt"

	// ===== 使用新 API =====
	// 获取 Object 服务
	objectSvc := client.Object()

	// 1. 上传对象（从字符串）
	log.Println("\n=== 上传对象（从字符串）===")
	data := strings.NewReader("Hello, RustFS! 这是一个测试对象。")
	uploadInfo, err := objectSvc.Put(ctx, bucketName, objectName, data, int64(data.Len()),
		object.WithContentType("text/plain; charset=utf-8"),
		object.WithUserMetadata(map[string]string{
			"author":  "rustfs-go",
			"version": "1.0",
		}),
		object.WithUserTags(map[string]string{
			"category": "example",
			"env":      "development",
		}),
	)
	if err != nil {
		log.Fatalln("上传对象失败:", err)
	}
	log.Printf("✅ 成功上传对象: %s\n", uploadInfo.Key)
	log.Printf("   ETag: %s\n", uploadInfo.ETag)
	log.Printf("   大小: %d bytes\n", uploadInfo.Size)
	if uploadInfo.VersionID != "" {
		log.Printf("   版本 ID: %s\n", uploadInfo.VersionID)
	}

	// 2. 获取对象信息
	log.Println("\n=== 获取对象信息 ===")
	objInfo, err := objectSvc.Stat(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalln("获取对象信息失败:", err)
	}
	log.Printf("对象: %s\n", objInfo.Key)
	log.Printf("  大小: %d bytes\n", objInfo.Size)
	log.Printf("  类型: %s\n", objInfo.ContentType)
	log.Printf("  ETag: %s\n", objInfo.ETag)
	log.Printf("  修改时间: %s\n", objInfo.LastModified.Format("2006-01-02 15:04:05"))
	if len(objInfo.UserMetadata) > 0 {
		log.Println("  用户元数据:")
		for k, v := range objInfo.UserMetadata {
			log.Printf("    %s: %s\n", k, v)
		}
	}

	// 3. 下载对象
	log.Println("\n=== 下载对象 ===")
	reader, _, err := objectSvc.Get(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalln("下载对象失败:", err)
	}
	defer reader.Close()

	buf := make([]byte, 1024)
	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		log.Fatalln("读取对象内容失败:", err)
	}
	log.Printf("对象内容: %s\n", string(buf[:n]))

	// 4. 下载对象的一部分（Range 请求）
	log.Println("\n=== 下载对象的一部分（Range 请求）===")
	rangeReader, _, err := objectSvc.Get(ctx, bucketName, objectName,
		object.WithGetRange(0, 10), // 只下载前 11 个字节（0-10）
	)
	if err != nil {
		log.Fatalln("Range 下载失败:", err)
	}
	defer rangeReader.Close()

	rangeBuf := make([]byte, 20)
	n, _ = rangeReader.Read(rangeBuf)
	log.Printf("部分内容（0-10 字节）: %s\n", string(rangeBuf[:n]))

	// 5. 列出对象
	log.Printf("\n=== 列出存储桶 %s 中的对象 ===\n", bucketName)
	objectsCh := objectSvc.List(ctx, bucketName)
	count := 0
	for obj := range objectsCh {
		if obj.Err != nil {
			log.Printf("列出对象时出错: %v\n", obj.Err)
			break
		}
		count++
		log.Printf("  %d. %s (大小: %d bytes)\n", count, obj.Key, obj.Size)
	}

	// 6. 复制对象
	log.Println("\n=== 复制对象 ===")
	copyObjectName := "test-object-copy.txt"
	copyInfo, err := objectSvc.Copy(ctx,
		bucketName, copyObjectName, // 目标
		bucketName, objectName, // 源
		object.WithCopyMetadata(map[string]string{
			"copied": "true",
		}, true), // 替换元数据
	)
	if err != nil {
		log.Printf("复制对象失败: %v\n", err)
	} else {
		log.Printf("✅ 成功复制对象: %s -> %s\n", objectName, copyObjectName)
		log.Printf("   新对象 ETag: %s\n", copyInfo.ETag)
	}

	// 7. 删除对象
	// log.Println("\n=== 删除对象 ===")
	// err = objectSvc.Delete(ctx, bucketName, objectName)
	// if err != nil {
	// 	log.Fatalln("删除对象失败:", err)
	// }
	// log.Printf("✅ 成功删除对象: %s\n", objectName)

	// // 删除复制的对象
	// err = objectSvc.Delete(ctx, bucketName, copyObjectName)
	// if err != nil {
	// 	log.Printf("删除复制对象失败: %v\n", err)
	// } else {
	// 	log.Printf("✅ 成功删除对象: %s\n", copyObjectName)
	// }

	log.Println("\n=== 示例运行完成 ===")
}
