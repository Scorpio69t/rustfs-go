//go:build example
// +build example

package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/internal/transport"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "mybucket"
	)

	// 初始化客户端
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln("初始化客户端失败:", err)
	}

	ctx := context.Background()

	// 示例 1: 基本的 HTTP 追踪
	fmt.Println("=== 示例 1: 基本的 HTTP 请求追踪 ===")
	traceBasicRequest(client, ctx, YOURBUCKET)

	// 示例 2: 追踪上传操作的性能
	fmt.Println("\n=== 示例 2: 追踪上传操作的性能 ===")
	traceUploadPerformance(client, ctx, YOURBUCKET)

	// 示例 3: 追踪列表操作
	fmt.Println("\n=== 示例 3: 追踪列表操作 ===")
	traceListOperation(client, ctx, YOURBUCKET)

	// 示例 4: 分析连接复用
	fmt.Println("\n=== 示例 4: 分析连接复用 ===")
	traceConnectionReuse(client, ctx, YOURBUCKET)
}

// traceBasicRequest 追踪基本请求
func traceBasicRequest(client *rustfs.Client, ctx context.Context, bucketName string) {
	var traceInfo *transport.TraceInfo

	// 创建带追踪的 context
	hook := func(info transport.TraceInfo) {
		// 保存追踪信息
		traceCopy := info
		traceInfo = &traceCopy
	}

	traceCtx := transport.NewTraceContext(ctx, hook)

	// 执行一个简单的桶存在性检查
	bucketSvc := client.Bucket()
	exists, err := bucketSvc.Exists(traceCtx, bucketName)
	if err != nil {
		log.Printf("检查存储桶失败: %v\n", err)
		return
	}

	fmt.Printf("存储桶 '%s' 存在: %v\n", bucketName, exists)

	if traceInfo != nil {
		fmt.Println("\n📊 追踪信息:")
		fmt.Printf("   连接复用: %v\n", traceInfo.ConnReused)
		fmt.Printf("   连接空闲: %v\n", traceInfo.ConnWasIdle)
		if traceInfo.ConnIdleTime > 0 {
			fmt.Printf("   空闲时长: %v\n", traceInfo.ConnIdleTime)
		}

		// 显示各阶段耗时
		timings := traceInfo.GetTimings()
		if len(timings) > 0 {
			fmt.Println("\n⏱️  各阶段耗时:")
			for stage, duration := range timings {
				fmt.Printf("   %s: %v\n", stage, duration)
			}
		}

		totalDuration := traceInfo.TotalDuration()
		if totalDuration > 0 {
			fmt.Printf("\n⏰ 总耗时: %v\n", totalDuration)
		}
	}
}

// traceUploadPerformance 追踪上传性能
func traceUploadPerformance(client *rustfs.Client, ctx context.Context, bucketName string) {
	// 准备测试数据
	testData := strings.Repeat("Hello, RustFS! ", 1000) // 约 15KB
	objectName := "trace-test-upload.txt"

	var uploadTrace *transport.TraceInfo

	hook := func(info transport.TraceInfo) {
		traceCopy := info
		uploadTrace = &traceCopy
	}

	traceCtx := transport.NewTraceContext(ctx, hook)

	// 上传对象
	objectSvc := client.Object()
	reader := strings.NewReader(testData)
	uploadInfo, err := objectSvc.Put(traceCtx, bucketName, objectName,
		reader, int64(len(testData)))
	if err != nil {
		log.Printf("上传失败: %v\n", err)
		return
	}

	fmt.Printf("✅ 上传成功: %s (ETag: %s)\n", objectName, uploadInfo.ETag)

	if uploadTrace != nil {
		fmt.Println("\n📊 上传性能分析:")
		fmt.Printf("   数据大小: %d 字节\n", len(testData))
		fmt.Printf("   连接复用: %v\n", uploadTrace.ConnReused)

		timings := uploadTrace.GetTimings()
		if requestWrite, ok := timings["request_write"]; ok {
			fmt.Printf("   写入请求耗时: %v\n", requestWrite)
		}
		if serverProcessing, ok := timings["server_processing"]; ok {
			fmt.Printf("   服务器处理耗时: %v\n", serverProcessing)
		}

		totalDuration := uploadTrace.TotalDuration()
		if totalDuration > 0 {
			// 计算上传速度
			speed := float64(len(testData)) / totalDuration.Seconds() / 1024 / 1024
			fmt.Printf("   总耗时: %v\n", totalDuration)
			fmt.Printf("   上传速度: %.2f MB/s\n", speed)
		}
	}
}

// traceListOperation 追踪列表操作
func traceListOperation(client *rustfs.Client, ctx context.Context, bucketName string) {
	var listTrace *transport.TraceInfo

	hook := func(info transport.TraceInfo) {
		traceCopy := info
		listTrace = &traceCopy
	}

	traceCtx := transport.NewTraceContext(ctx, hook)

	// 列出对象
	objectSvc := client.Object()
	objectsCh := objectSvc.List(traceCtx, bucketName)

	count := 0
	for obj := range objectsCh {
		if obj.Err != nil {
			log.Printf("列表错误: %v\n", obj.Err)
			break
		}
		count++
		if count <= 5 { // 只显示前 5 个
			fmt.Printf("   - %s (%d bytes)\n", obj.Key, obj.Size)
		}
	}

	if count > 5 {
		fmt.Printf("   ... 还有 %d 个对象\n", count-5)
	}

	fmt.Printf("\n总共: %d 个对象\n", count)

	if listTrace != nil {
		fmt.Println("\n📊 列表操作性能:")
		fmt.Printf("   连接复用: %v\n", listTrace.ConnReused)

		timings := listTrace.GetTimings()
		if serverProcessing, ok := timings["server_processing"]; ok {
			fmt.Printf("   服务器处理耗时: %v\n", serverProcessing)
		}

		totalDuration := listTrace.TotalDuration()
		if totalDuration > 0 {
			fmt.Printf("   总耗时: %v\n", totalDuration)
			if count > 0 {
				avgTime := totalDuration.Microseconds() / int64(count)
				fmt.Printf("   平均每个对象: %d μs\n", avgTime)
			}
		}
	}
}

// traceConnectionReuse 分析连接复用
func traceConnectionReuse(client *rustfs.Client, ctx context.Context, bucketName string) {
	fmt.Println("执行 5 次连续请求，观察连接复用情况...\n")

	bucketSvc := client.Bucket()

	for i := 1; i <= 5; i++ {
		var traceInfo *transport.TraceInfo

		hook := func(info transport.TraceInfo) {
			traceCopy := info
			traceInfo = &traceCopy
		}

		traceCtx := transport.NewTraceContext(ctx, hook)

		// 执行请求
		_, err := bucketSvc.Exists(traceCtx, bucketName)
		if err != nil {
			log.Printf("请求 %d 失败: %v\n", i, err)
			continue
		}

		if traceInfo != nil {
			status := "🆕 新连接"
			if traceInfo.ConnReused {
				status = "♻️  复用连接"
				if traceInfo.ConnWasIdle {
					status += fmt.Sprintf(" (空闲了 %v)", traceInfo.ConnIdleTime)
				}
			}

			totalDuration := traceInfo.TotalDuration()
			fmt.Printf("请求 %d: %s - 耗时: %v\n", i, status, totalDuration)

			// 第一次请求显示详细的建立连接时间
			if i == 1 && !traceInfo.ConnReused {
				timings := traceInfo.GetTimings()
				if dnsLookup, ok := timings["dns_lookup"]; ok {
					fmt.Printf("         DNS 查询: %v\n", dnsLookup)
				}
				if tcpConnect, ok := timings["tcp_connect"]; ok {
					fmt.Printf("         TCP 连接: %v\n", tcpConnect)
				}
			}
		}
	}

	fmt.Println("\n💡 提示:")
	fmt.Println("   - 新连接需要 DNS 查询和 TCP 握手，耗时较长")
	fmt.Println("   - 复用连接可以显著提高性能")
	fmt.Println("   - SDK 自动管理连接池，无需手动处理")
}
