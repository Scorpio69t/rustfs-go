//go:build example
// +build example

// presigned.go - 预签名 URL 示例（使用旧 API）
// 注意：预签名 URL 功能暂未迁移到新 API，此示例仍使用旧 API
package main

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "mybucket" // 'mc mb play/mybucket' if it does not exist.
	)

	// 初始化客户端
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	bucketName := YOURBUCKET
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
	policy := rustfs.NewPostPolicy()
	err = policy.SetExpires(time.Now().Add(time.Hour))
	if err != nil {
		log.Fatalln(err)
		return
	}

	err = policy.SetCondition("$eq", "bucket", bucketName)
	if err != nil {
		log.Fatalln(err)
		return
	}

	err = policy.SetCondition("$eq", "key", "post-object.txt")
	if err != nil {
		log.Fatalln(err)
		return
	}

	err = policy.SetCondition("$eq", "Content-Type", "text/plain")
	if err != nil {
		log.Fatalln(err)
		return
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
