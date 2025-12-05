//go:build example
// +build example

package main

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/credentials"
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
	objectName := "test-object.txt"

	// 生成预签名 GET URL（1小时有效）
	presignedURL, err := client.PresignedGetObject(ctx, bucketName, objectName, time.Hour, url.Values{})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("预签名 GET URL (1小时有效):\n%s\n", presignedURL.String())

	// 生成预签名 PUT URL（1小时有效）
	presignedPutURL, err := client.PresignedPutObject(ctx, bucketName, "upload-object.txt", time.Hour)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("预签名 PUT URL (1小时有效):\n%s\n", presignedPutURL.String())

	// 生成预签名 POST URL
	policy := &rustfs.PostPolicy{
		Expiration: time.Now().Add(time.Hour),
		Conditions: []map[string]interface{}{
			{"bucket": bucketName},
			{"key": "post-object.txt"},
			{"Content-Type": "text/plain"},
		},
	}
	postURL, formData, err := client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("预签名 POST URL:\n%s\n", postURL.String())
	log.Println("表单数据:")
	for k, v := range formData {
		log.Printf("  %s: %s\n", k, v)
	}
}
