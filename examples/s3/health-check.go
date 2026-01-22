//go:build example
// +build example

// Example: Service health check
// Demonstrates how to check the health status of a RustFS service
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
	accessKey = "rustfsadmin"
	secretKey = "rustfsadmin"
)

func main() {
	// Create client
	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("Running RustFS service health checks...")
	fmt.Println(strings.Repeat("=", 50))

	// 1. Basic health check
	fmt.Println("\n1) Basic health check")
	result := client.HealthCheck(nil)
	printHealthResult(result)

	// 2. Health check with timeout
	fmt.Println("\n2) Health check with timeout (5s)")
	opts := &core.HealthCheckOptions{
		Timeout: 5 * time.Second,
		Context: context.Background(),
	}
	result = client.HealthCheck(opts)
	printHealthResult(result)

	// 3. Health check with retry
	fmt.Println("\n3) Health check with retry (up to 3 attempts)")
	result = client.HealthCheckWithRetry(opts, 3)
	printHealthResult(result)

	// 4. Continuous monitoring (demo)
	fmt.Println("\n4) Continuous monitoring (every 5s, 3 checks)")
	for i := 1; i <= 3; i++ {
		fmt.Printf("\nCheck #%d:\n", i)
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
		fmt.Printf("✅ Service healthy\n")
		fmt.Printf("   Response time: %v\n", result.ResponseTime)
		fmt.Printf("   Checked at: %s\n", result.CheckedAt.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("❌ Service unhealthy\n")
		fmt.Printf("   Error: %v\n", result.Error)
		fmt.Printf("   Response time: %v\n", result.ResponseTime)
		fmt.Printf("   Checked at: %s\n", result.CheckedAt.Format("2006-01-02 15:04:05"))
	}
}
