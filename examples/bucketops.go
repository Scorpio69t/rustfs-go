//go:build example
// +build example

package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// 初始化客户端
	endpoint := "127.0.0.1:9000"
	accessKeyID := "rustfsadmin"
	secretAccessKey := "rustfsadmin"

	client, err := rustfs.New(endpoint, &rustfs.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
		Region: "us-east-1",
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()

	// 创建存储桶
	bucketName := "test-bucket"
	err = client.MakeBucket(ctx, bucketName, rustfs.MakeBucketOptions{
		Region: "us-east-1",
	})
	if err != nil {
		log.Printf("创建存储桶失败: %v\n", err)
	} else {
		log.Printf("成功创建存储桶: %s\n", bucketName)
	}

	// 检查存储桶是否存在
	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("存储桶 %s 是否存在: %v\n", bucketName, exists)

	// 列出所有存储桶
	buckets, err := client.ListBuckets(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("存储桶列表:")
	for _, bucket := range buckets {
		log.Printf("  - %s (创建时间: %s)\n", bucket.Name, bucket.CreationDate.Format("2006-01-02 15:04:05"))
	}

	// 列出存储桶中的对象
	log.Printf("\n存储桶 %s 中的对象:\n", bucketName)
	objectsCh := client.ListObjects(ctx, bucketName, rustfs.ListObjectsOptions{
		Prefix:  "",
		MaxKeys: 100,
	})
	for obj := range objectsCh {
		log.Printf("  - %s (大小: %d bytes, 修改时间: %s)\n",
			obj.Key, obj.Size, obj.LastModified.Format("2006-01-02 15:04:05"))
	}

	// 删除存储桶（需要先清空存储桶中的对象）
	// err = client.RemoveBucket(ctx, bucketName, rustfs.RemoveBucketOptions{})
	// if err != nil {
	// 	log.Printf("删除存储桶失败: %v\n", err)
	// } else {
	// 	log.Printf("成功删除存储桶: %s\n", bucketName)
	// }
}
