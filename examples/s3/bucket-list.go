//go:build example
// +build example

// Example: List all buckets
// Demonstrates how to use the RustFS Go SDK to list all buckets in the account
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
	)

	// 初始化 RustFS 客户端
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Unable to create client: %v", err)
	}

	ctx := context.Background()

	// 获取 Bucket 服务
	bucketSvc := client.Bucket()

	// List all buckets
	buckets, err := bucketSvc.List(ctx)
	if err != nil {
		log.Fatalf("Failed to list buckets: %v", err)
	}

	// Show results
	log.Printf("Found %d buckets:", len(buckets))
	log.Println("----------------------------------------")

	for i, bucket := range buckets {
		log.Printf("%d. Name: %s", i+1, bucket.Name)
		log.Printf("   Created: %s", bucket.CreationDate.Format("2006-01-02 15:04:05"))
		if bucket.Region != "" {
			log.Printf("   Region: %s", bucket.Region)
		}
		log.Println("----------------------------------------")
	}

	if len(buckets) == 0 {
		log.Println("No buckets found")
	}
}
