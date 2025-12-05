//go:build example
// +build example

package main

import (
	"context"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go/v1"
	"github.com/Scorpio69t/rustfs-go/v1/credentials"
)

func main() {
	// 初始化客户端
	endpoint := "rustfs.example.com"
	accessKeyID := "your-access-key"
	secretAccessKey := "your-secret-key"

	client, err := rustfs.New(endpoint, &rustfs.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: true,
		Region: "us-east-1",
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	bucketName := "test-bucket"
	objectName := "large-file.txt"

	// 初始化多部分上传
	uploadID, err := client.InitiateMultipartUpload(ctx, bucketName, objectName, rustfs.PutObjectOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("初始化多部分上传成功，UploadID: %s\n", uploadID)

	// 准备数据（模拟大文件分片）
	parts := []rustfs.CompletePart{}
	partSize := int64(5 * 1024 * 1024) // 5MB per part

	// 上传第一个分片
	part1Data := strings.NewReader(strings.Repeat("A", int(partSize)))
	part1, err := client.UploadPart(ctx, bucketName, objectName, uploadID, 1, part1Data, partSize, rustfs.PutObjectPartOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("上传分片 1 成功，ETag: %s\n", part1.ETag)
	parts = append(parts, rustfs.CompletePart{
		PartNumber: part1.PartNumber,
		ETag:       part1.ETag,
	})

	// 上传第二个分片
	part2Data := strings.NewReader(strings.Repeat("B", int(partSize)))
	part2, err := client.UploadPart(ctx, bucketName, objectName, uploadID, 2, part2Data, partSize, rustfs.PutObjectPartOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("上传分片 2 成功，ETag: %s\n", part2.ETag)
	parts = append(parts, rustfs.CompletePart{
		PartNumber: part2.PartNumber,
		ETag:       part2.ETag,
	})

	// 完成多部分上传
	uploadInfo, err := client.CompleteMultipartUpload(ctx, bucketName, objectName, uploadID, parts, rustfs.PutObjectOptions{})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("完成多部分上传成功，ETag: %s\n", uploadInfo.ETag)

	// 如果需要取消上传，可以使用：
	// err = client.AbortMultipartUpload(ctx, bucketName, objectName, uploadID, rustfs.AbortMultipartUploadOptions{})
	// if err != nil {
	// 	log.Fatalln(err)
	// }
	// log.Println("取消多部分上传成功")
}
