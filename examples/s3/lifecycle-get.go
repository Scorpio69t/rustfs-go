//go:build example
// +build example

// Example: Get bucket lifecycle policy
// Demonstrates how to retrieve the current lifecycle configuration
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

const (
	endpoint  = "127.0.0.1:9000"
	accessKey = "XhJOoEKn3BM6cjD2dVmx"
	secretKey = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
	bucket    = "mybucket"
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

	ctx := context.Background()
	service := client.Bucket()

	fmt.Printf("Retrieving lifecycle configuration for bucket '%s'...\n\n", bucket)

	// Get lifecycle configuration
	config, err := service.GetLifecycle(ctx, bucket)
	if err != nil {
		log.Fatalf("Failed to get lifecycle configuration: %v\n", err)
	}

	if len(config) == 0 {
		fmt.Println("No lifecycle configuration is set for this bucket")
		return
	}

	fmt.Println("✅ Lifecycle configuration:")
	fmt.Println(string(config))
}
