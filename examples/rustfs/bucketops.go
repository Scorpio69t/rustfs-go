//go:build example
// +build example

// bucketops-new.go - 使用新 API 的存储桶操作示例
package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/bucket"
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

	// ===== 使用新 API =====
	// 获取 Bucket 服务
	bucketSvc := client.Bucket()

	// 1. 创建存储桶
	log.Println("\n=== 创建存储桶 ===")
	err = bucketSvc.Create(ctx, bucketName,
		bucket.WithRegion("us-east-1"),
		// bucket.WithObjectLocking(true), // 可选：启用对象锁定
	)
	if err != nil {
		log.Printf("创建存储桶失败: %v\n", err)
	} else {
		log.Printf("✅ 成功创建存储桶: %s\n", bucketName)
	}

	// 2. 检查存储桶是否存在
	log.Println("\n=== 检查存储桶是否存在 ===")
	exists, err := bucketSvc.Exists(ctx, bucketName)
	if err != nil {
		log.Fatalln("检查存储桶失败:", err)
	}
	log.Printf("存储桶 %s 是否存在: %v\n", bucketName, exists)

	// 3. 获取存储桶位置
	log.Println("\n=== 获取存储桶位置 ===")
	location, err := bucketSvc.GetLocation(ctx, bucketName)
	if err != nil {
		log.Fatalln("获取存储桶位置失败:", err)
	}
	log.Printf("存储桶 %s 的位置: %s\n", bucketName, location)

	// 4. 列出所有存储桶
	log.Println("\n=== 列出所有存储桶 ===")
	buckets, err := bucketSvc.List(ctx)
	if err != nil {
		log.Fatalln("列出存储桶失败:", err)
	}
	log.Printf("共有 %d 个存储桶:\n", len(buckets))
	for i, b := range buckets {
		log.Printf("  %d. %s (创建时间: %s)\n",
			i+1, b.Name, b.CreationDate.Format("2006-01-02 15:04:05"))
	}

	// 5. 列出存储桶中的对象（使用 Object 服务）
	log.Printf("\n=== 存储桶 %s 中的对象 ===\n", bucketName)
	objectSvc := client.Object()
	objectsCh := objectSvc.List(ctx, bucketName)
	count := 0
	for obj := range objectsCh {
		if obj.Err != nil {
			log.Printf("列出对象时出错: %v\n", obj.Err)
			break
		}
		count++
		log.Printf("  %d. %s (大小: %d bytes, 修改时间: %s)\n",
			count, obj.Key, obj.Size, obj.LastModified.Format("2006-01-02 15:04:05"))
	}
	if count == 0 {
		log.Println("  存储桶为空")
	}

	// 6. 删除存储桶（需要先清空存储桶中的对象）
	// log.Println("\n=== 删除存储桶 ===")
	// err = bucketSvc.Delete(ctx, bucketName)

	// if err != nil {
	// 	log.Printf("删除存储桶失败: %v\n", err)
	// } else {
	// 	log.Printf("✅ 成功删除存储桶: %s\n", bucketName)
	// }

	log.Println("\n=== 示例运行完成 ===")
}
