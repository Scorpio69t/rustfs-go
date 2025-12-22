//go:build example

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
		YOURBUCKET          = "mybucket"
	)

	// Initialize RustFS client
	client, err := rustfs.New("127.0.0.1:9000", &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln("Initialize RustFS client failed:", err)
	}

	ctx := context.Background()

	// Example 1: Basic health check
	fmt.Println("=== Example 1: Basic health check ===")
	healthCheck(client)

	// Example 2: Health check with timeout
	fmt.Println("\n=== Example 2: Health check with timeout ===")
	healthCheckWithTimeout(client)

	// Example 3: Check specific bucket health
	fmt.Println("\n=== Example 3: Check specific bucket health ===")
	checkBucketHealth(client, ctx)

	// Example 4: Health check with retries
	fmt.Println("\n=== Example 4: Health check with retries ===")
	healthCheckWithRetry(client)

	// Example 5: Periodic health check
	fmt.Println("\n=== Example 5: Periodic health check ===")
	periodicHealthCheck(client)
}

// health check execute a basic health check
func healthCheck(client *rustfs.Client) {
	result := client.HealthCheck(nil)

	if result.Healthy {
		fmt.Printf("✅ Service is healthy\n")
		fmt.Printf("   Endpoint: %s\n", result.Endpoint)
		fmt.Printf("   Region: %s\n", result.Region)
		fmt.Printf("   ResponseTime: %v\n", result.ResponseTime)
		fmt.Printf("   StatusCode: %d\n", result.StatusCode)
	} else {
		fmt.Printf("❌ Service is unhealthy\n")
		fmt.Printf("   Error: %v\n", result.Error)
	}
}

// healthCheckWithTimeout withTimeout performs a health check with a specified timeout
func healthCheckWithTimeout(client *rustfs.Client) {
	opts := &core.HealthCheckOptions{
		Timeout: 2 * time.Second,
		Context: context.Background(),
	}

	result := client.HealthCheck(opts)
	fmt.Printf("Health status: %s\n", result.String())
}

// checkBucketHealth checks the health of a specific bucket
func checkBucketHealth(client *rustfs.Client, ctx context.Context) {
	bucketName := "mybucket"

	// create the bucket if it does not exist
	bucketSvc := client.Bucket()
	exists, _ := bucketSvc.Exists(ctx, bucketName)

	if !exists {
		fmt.Printf("Creating bucket '%s'...\n", bucketName)
		if err := bucketSvc.Create(ctx, bucketName); err != nil {
			log.Printf("Failed to create bucket '%s': %v\n", bucketName, err)
			return
		}
	}

	// create health check options
	opts := &core.HealthCheckOptions{
		Timeout:    3 * time.Second,
		BucketName: bucketName,
		Context:    ctx,
	}

	result := client.HealthCheck(opts)

	if result.Healthy {
		fmt.Printf("✅ Bucket '%s' is healthy\n", bucketName)
		fmt.Printf("   ResponseTime: %v\n", result.ResponseTime)
	} else {
		fmt.Printf("❌ Bucket '%s' is unhealthy\n", bucketName)
		fmt.Printf("   Error: %v\n", result.Error)
	}
}

// healthCheckWithRetry performs a health check with retries
func healthCheckWithRetry(client *rustfs.Client) {
	opts := &core.HealthCheckOptions{
		Timeout: 3 * time.Second,
		Context: context.Background(),
	}

	fmt.Println("Performing health check with retries...")
	result := client.HealthCheckWithRetry(opts, 3)

	if result == nil {
		fmt.Println("❌ Health check failed after retries")
		return
	}

	if result.Healthy {
		fmt.Printf("✅ Service is healthy\n")
		fmt.Printf("   ResponseTime: %v\n", result.ResponseTime)
	} else {
		fmt.Printf("❌ Service is unhealthy\n")
		fmt.Printf("   Error: %v\n", result.Error)
	}
}

// periodicHealthCheck performs periodic health checks every 5 seconds
func periodicHealthCheck(client *rustfs.Client) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	opts := &core.HealthCheckOptions{
		Timeout: 2 * time.Second,
		Context: context.Background(),
	}

	// Perform health checks 3 times
	count := 0
	for range ticker.C {
		count++
		result := client.HealthCheck(opts)

		timestamp := time.Now().Format("15:04:05")
		if result.Healthy {
			fmt.Printf("[%s] ✅ Healthy - ResponseTime: %v\n", timestamp, result.ResponseTime)
		} else {
			fmt.Printf("[%s] ❌ Unhealthy - Error: %v\n", timestamp, result.Error)
		}

		if count >= 3 {
			fmt.Println("Stopping periodic health checks.")
			break
		}
	}
}
