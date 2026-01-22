//go:build example
// +build example

// 示例：设置存储桶策略
// 演示如何使用 RustFS Go SDK 设置存储桶访问策略
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
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
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

	// 获取 Bucket 服务
	bucketSvc := client.Bucket()

	bucketName := YOURBUCKET

	// 定义存储桶策略（JSON 格式）
	// 这个策略允许匿名用户读取指定前缀的对象
	policy := `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {"AWS": "*"},
            "Action": ["s3:GetObject"],
            "Resource": ["arn:aws:s3:::` + bucketName + `/public/*"]
        }
    ]
}`

	// 设置存储桶策略
	err = bucketSvc.SetPolicy(ctx, bucketName, policy)
	if err != nil {
		log.Fatalf("设置存储桶策略失败: %v", err)
	}

	log.Printf("✅ 存储桶 '%s' 策略设置成功", bucketName)
	log.Println("\n已应用的策略:")
	log.Println("----------------------------------------")
	log.Println(policy)
	log.Println("----------------------------------------")
	log.Println("\n提示：此策略允许公开读取 'public/' 前缀下的对象")
	log.Println("使用 policy-get.go 查看当前策略")
}
