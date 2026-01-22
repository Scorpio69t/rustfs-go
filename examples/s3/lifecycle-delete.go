//go:build example
// +build example

// Example: Delete bucket lifecycle policy
// Demonstrates how to remove a bucket's lifecycle configuration
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
	accessKey = "rustfsadmin"
	secretKey = "rustfsadmin"
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

	fmt.Printf("Deleting lifecycle configuration for bucket '%s'...\n", bucket)

	// Delete lifecycle configuration
	err = service.DeleteLifecycle(ctx, bucket)
	if err != nil {
		log.Fatalf("Failed to delete lifecycle configuration: %v\n", err)
	}

	fmt.Println("✅ Lifecycle configuration deleted")
}
