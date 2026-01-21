//go:build example
// +build example

// 示例：服务健康检查
// 演示如何检查 RustFS 服务的健康状态
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

const (
	endpoint  = "127.0.0.1:9000"
	accessKey = "XhJOoEKn3BM6cjD2dVmx"
	secretKey = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
)

func main() {
	// 创建客户端
	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("执行 RustFS 服务健康检查...")
	fmt.Println(strings.Repeat("=", 50))

	// 1. 基本健康检查
	fmt.Println("\n1️⃣ 基本健康检查")
	result := client.HealthCheck(nil)
	printHealthResult(result)

	// 2. 带超时的健康检查
	fmt.Println("\n2️⃣ 带超时的健康检查 (5秒超时)")
	opts := &core.HealthCheckOptions{
		Timeout: 5 * time.Second,
		Context: context.Background(),
	}
	result = client.HealthCheck(opts)
	printHealthResult(result)

	// 3. 带重试的健康检查
	fmt.Println("\n3️⃣ 带重试的健康检查 (最多3次)")
	result = client.HealthCheckWithRetry(opts, 3)
	printHealthResult(result)

	// 4. 连续监控（演示）
	fmt.Println("\n4️⃣ 连续监控 (每5秒检查一次，共3次)")
	for i := 1; i <= 3; i++ {
		fmt.Printf("\n检查 #%d:\n", i)
		result = client.HealthCheck(opts)
		printHealthResult(result)
		if i < 3 {
			time.Sleep(5 * time.Second)
		}
	}

	fmt.Println(strings.Repeat("=", 50))
}

func printHealthResult(result *core.HealthCheckResult) {
	if result.Healthy {
		fmt.Printf("✅ 服务健康\n")
		fmt.Printf("   响应时间: %v\n", result.ResponseTime)
		fmt.Printf("   检查时间: %s\n", result.CheckedAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("❌ 服务不健康\n")
		fmt.Printf("   错误信息: %v\n", result.Error)
		fmt.Printf("   响应时间: %v\n", result.ResponseTime)
		fmt.Printf("   检查时间: %s\n", result.CheckedAt.Format("2006-01-02 15:04:05"))
	}
}
