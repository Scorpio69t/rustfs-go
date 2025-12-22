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

	// Initialize client
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln("Failed to initialize client:", err)
	}

	ctx := context.Background()

	// Example 1: Basic HTTP tracing
	fmt.Println("=== Example 1: Basic HTTP request tracing ===")
	traceBasicRequest(client, ctx, YOURBUCKET)

	// Example 2: Trace upload operation performance
	fmt.Println("\n=== Example 2: Trace upload operation performance ===")
	traceUploadPerformance(client, ctx, YOURBUCKET)

	// Example 3: Trace list operation
	fmt.Println("\n=== Example 3: Trace list operation ===")
	traceListOperation(client, ctx, YOURBUCKET)

	// Example 4: Analyze connection reuse
	fmt.Println("\n=== Example 4: Analyze connection reuse ===")
	traceConnectionReuse(client, ctx, YOURBUCKET)
}

// traceBasicRequest traces basic request
func traceBasicRequest(client *rustfs.Client, ctx context.Context, bucketName string) {
	var traceInfo *transport.TraceInfo

	// Create context with tracing
	hook := func(info transport.TraceInfo) {
		// Save trace information
		traceCopy := info
		traceInfo = &traceCopy
	}

	traceCtx := transport.NewTraceContext(ctx, hook)

	// Execute a simple bucket existence check
	bucketSvc := client.Bucket()
	exists, err := bucketSvc.Exists(traceCtx, bucketName)
	if err != nil {
		log.Printf("Failed to check bucket: %v\n", err)
		return
	}

	fmt.Printf("Bucket '%s' exists: %v\n", bucketName, exists)

	if traceInfo != nil {
		fmt.Println("\n📊 Trace information:")
		fmt.Printf("   Connection reused: %v\n", traceInfo.ConnReused)
		fmt.Printf("   Connection was idle: %v\n", traceInfo.ConnWasIdle)
		if traceInfo.ConnIdleTime > 0 {
			fmt.Printf("   Idle duration: %v\n", traceInfo.ConnIdleTime)
		}

		// Display timing for each stage
		timings := traceInfo.GetTimings()
		if len(timings) > 0 {
			fmt.Println("\n⏱️  Stage timings:")
			for stage, duration := range timings {
				fmt.Printf("   %s: %v\n", stage, duration)
			}
		}

		totalDuration := traceInfo.TotalDuration()
		if totalDuration > 0 {
			fmt.Printf("\n⏰ Total duration: %v\n", totalDuration)
		}
	}
}

// traceUploadPerformance traces upload performance
func traceUploadPerformance(client *rustfs.Client, ctx context.Context, bucketName string) {
	// Prepare test data
	testData := strings.Repeat("Hello, RustFS! ", 1000) // ~15KB
	objectName := "trace-test-upload.txt"

	var uploadTrace *transport.TraceInfo

	hook := func(info transport.TraceInfo) {
		traceCopy := info
		uploadTrace = &traceCopy
	}

	traceCtx := transport.NewTraceContext(ctx, hook)

	// Upload object
	objectSvc := client.Object()
	reader := strings.NewReader(testData)
	uploadInfo, err := objectSvc.Put(traceCtx, bucketName, objectName,
		reader, int64(len(testData)))
	if err != nil {
		log.Printf("Upload failed: %v\n", err)
		return
	}

	fmt.Printf("✅ Upload successful: %s (ETag: %s)\n", objectName, uploadInfo.ETag)

	if uploadTrace != nil {
		fmt.Println("\n📊 Upload performance analysis:")
		fmt.Printf("   Data size: %d bytes\n", len(testData))
		fmt.Printf("   Connection reused: %v\n", uploadTrace.ConnReused)

		timings := uploadTrace.GetTimings()
		if requestWrite, ok := timings["request_write"]; ok {
			fmt.Printf("   Request write duration: %v\n", requestWrite)
		}
		if serverProcessing, ok := timings["server_processing"]; ok {
			fmt.Printf("   Server processing duration: %v\n", serverProcessing)
		}

		totalDuration := uploadTrace.TotalDuration()
		if totalDuration > 0 {
			// Calculate upload speed
			speed := float64(len(testData)) / totalDuration.Seconds() / 1024 / 1024
			fmt.Printf("   Total duration: %v\n", totalDuration)
			fmt.Printf("   Upload speed: %.2f MB/s\n", speed)
		}
	}
}

// traceListOperation traces list operation
func traceListOperation(client *rustfs.Client, ctx context.Context, bucketName string) {
	var listTrace *transport.TraceInfo

	hook := func(info transport.TraceInfo) {
		traceCopy := info
		listTrace = &traceCopy
	}

	traceCtx := transport.NewTraceContext(ctx, hook)

	// List objects
	objectSvc := client.Object()
	objectsCh := objectSvc.List(traceCtx, bucketName)

	count := 0
	for obj := range objectsCh {
		if obj.Err != nil {
			log.Printf("List error: %v\n", obj.Err)
			break
		}
		count++
		if count <= 5 { // Only show first 5
			fmt.Printf("   - %s (%d bytes)\n", obj.Key, obj.Size)
		}
	}

	if count > 5 {
		fmt.Printf("   ... %d more objects\n", count-5)
	}

	fmt.Printf("\nTotal: %d objects\n", count)

	if listTrace != nil {
		fmt.Println("\n📊 List operation performance:")
		fmt.Printf("   Connection reused: %v\n", listTrace.ConnReused)

		timings := listTrace.GetTimings()
		if serverProcessing, ok := timings["server_processing"]; ok {
			fmt.Printf("   Server processing duration: %v\n", serverProcessing)
		}

		totalDuration := listTrace.TotalDuration()
		if totalDuration > 0 {
			fmt.Printf("   Total duration: %v\n", totalDuration)
			if count > 0 {
				avgTime := totalDuration.Microseconds() / int64(count)
				fmt.Printf("   Average per object: %d μs\n", avgTime)
			}
		}
	}
}

// traceConnectionReuse analyzes connection reuse
func traceConnectionReuse(client *rustfs.Client, ctx context.Context, bucketName string) {
	fmt.Println("Executing 5 consecutive requests to observe connection reuse...\n")

	bucketSvc := client.Bucket()

	for i := 1; i <= 5; i++ {
		var traceInfo *transport.TraceInfo

		hook := func(info transport.TraceInfo) {
			traceCopy := info
			traceInfo = &traceCopy
		}

		traceCtx := transport.NewTraceContext(ctx, hook)

		// Execute request
		_, err := bucketSvc.Exists(traceCtx, bucketName)
		if err != nil {
			log.Printf("Request %d failed: %v\n", i, err)
			continue
		}

		if traceInfo != nil {
			status := "🆕 New connection"
			if traceInfo.ConnReused {
				status = "♻️  Reused connection"
				if traceInfo.ConnWasIdle {
					status += fmt.Sprintf(" (idle for %v)", traceInfo.ConnIdleTime)
				}
			}

			totalDuration := traceInfo.TotalDuration()
			fmt.Printf("Request %d: %s - Duration: %v\n", i, status, totalDuration)

			// First request shows detailed connection establishment time
			if i == 1 && !traceInfo.ConnReused {
				timings := traceInfo.GetTimings()
				if dnsLookup, ok := timings["dns_lookup"]; ok {
					fmt.Printf("         DNS lookup: %v\n", dnsLookup)
				}
				if tcpConnect, ok := timings["tcp_connect"]; ok {
					fmt.Printf("         TCP connect: %v\n", tcpConnect)
				}
			}
		}
	}

	fmt.Println("\n💡 Tips:")
	fmt.Println("   - New connections require DNS lookup and TCP handshake, taking longer")
	fmt.Println("   - Reusing connections can significantly improve performance")
	fmt.Println("   - SDK automatically manages connection pool, no manual handling needed")
}
