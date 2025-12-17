//go:build example
// +build example

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/internal/core"
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
	client, err := rustfs.New("127.0.0.1:9000", &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln("初始化客户端失败:", err)
	}

	ctx := context.Background()

	// 示例 1: 基本健康检查
	fmt.Println("=== 示例 1: 基本健康检查 ===")
	healthCheck(client)

	// 示例 2: 带超时的健康检查
	fmt.Println("\n=== 示例 2: 带超时的健康检查 ===")
	healthCheckWithTimeout(client)

	// 示例 3: 检查特定存储桶
	fmt.Println("\n=== 示例 3: 检查特定存储桶 ===")
	checkBucketHealth(client, ctx)

	// 示例 4: 带重试的健康检查
	fmt.Println("\n=== 示例 4: 带重试的健康检查 ===")
	healthCheckWithRetry(client)

	// 示例 5: 定期健康检查
	fmt.Println("\n=== 示例 5: 定期健康检查 (每 5 秒) ===")
	periodicHealthCheck(client)
}

// healthCheck 执行基本的健康检查
func healthCheck(client *rustfs.Client) {
	result := client.HealthCheck(nil)

	if result.Healthy {
		fmt.Printf("✅ 服务健康\n")
		fmt.Printf("   端点: %s\n", result.Endpoint)
		fmt.Printf("   区域: %s\n", result.Region)
		fmt.Printf("   响应时间: %v\n", result.ResponseTime)
		fmt.Printf("   状态码: %d\n", result.StatusCode)
	} else {
		fmt.Printf("❌ 服务不健康\n")
		fmt.Printf("   错误: %v\n", result.Error)
	}
}

// healthCheckWithTimeout 带自定义超时的健康检查
func healthCheckWithTimeout(client *rustfs.Client) {
	opts := &core.HealthCheckOptions{
		Timeout: 2 * time.Second,
		Context: context.Background(),
	}

	result := client.HealthCheck(opts)
	fmt.Printf("健康状态: %s\n", result.String())
}

// checkBucketHealth 检查特定存储桶的健康状态
func checkBucketHealth(client *rustfs.Client, ctx context.Context) {
	bucketName := "mybucket"

	// 先创建存储桶（如果不存在）
	bucketSvc := client.Bucket()
	exists, _ := bucketSvc.Exists(ctx, bucketName)

	if !exists {
		fmt.Printf("创建测试存储桶: %s\n", bucketName)
		if err := bucketSvc.Create(ctx, bucketName); err != nil {
			log.Printf("创建存储桶失败: %v\n", err)
			return
		}
	}

	// 检查存储桶健康状态
	opts := &core.HealthCheckOptions{
		Timeout:    3 * time.Second,
		BucketName: bucketName,
		Context:    ctx,
	}

	result := client.HealthCheck(opts)

	if result.Healthy {
		fmt.Printf("✅ 存储桶 '%s' 健康\n", bucketName)
		fmt.Printf("   响应时间: %v\n", result.ResponseTime)
	} else {
		fmt.Printf("❌ 存储桶 '%s' 不健康\n", bucketName)
		fmt.Printf("   错误: %v\n", result.Error)
	}
}

// healthCheckWithRetry 带重试的健康检查
func healthCheckWithRetry(client *rustfs.Client) {
	opts := &core.HealthCheckOptions{
		Timeout: 3 * time.Second,
		Context: context.Background(),
	}

	fmt.Println("执行健康检查（最多重试 3 次）...")
	result := client.HealthCheckWithRetry(opts, 3)

	if result.Healthy {
		fmt.Printf("✅ 服务健康（经过重试）\n")
		fmt.Printf("   响应时间: %v\n", result.ResponseTime)
	} else {
		fmt.Printf("❌ 服务不健康（重试后仍失败）\n")
		fmt.Printf("   错误: %v\n", result.Error)
	}
}

// periodicHealthCheck 定期执行健康检查
func periodicHealthCheck(client *rustfs.Client) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	opts := &core.HealthCheckOptions{
		Timeout: 2 * time.Second,
		Context: context.Background(),
	}

	// 执行 3 次检查后退出（演示用）
	count := 0
	for range ticker.C {
		count++
		result := client.HealthCheck(opts)

		timestamp := time.Now().Format("15:04:05")
		if result.Healthy {
			fmt.Printf("[%s] ✅ 健康 - 响应时间: %v\n", timestamp, result.ResponseTime)
		} else {
			fmt.Printf("[%s] ❌ 不健康 - 错误: %v\n", timestamp, result.Error)
		}

		if count >= 3 {
			fmt.Println("健康检查演示完成")
			break
		}
	}
}
