//go:build example
// +build example

// Example: Get bucket object lock configuration
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/objectlock"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "object-lock-bucket"
	)

	// Initialize RustFS client
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	bucketSvc := client.Bucket()

	config, err := bucketSvc.GetObjectLockConfig(ctx, YOURBUCKET)
	if err != nil {
		if errors.Is(err, objectlock.ErrNoObjectLockConfig) {
			log.Printf("No object lock configuration found for %s", YOURBUCKET)
			return
		}
		log.Fatalf("Failed to get object lock configuration: %v", err)
	}

	fmt.Printf("Object lock configuration for %s\n", YOURBUCKET)
	fmt.Printf("  Enabled: %s\n", config.ObjectLockEnabled)

	if config.Rule == nil {
		fmt.Println("  Default retention: none")
		return
	}

	fmt.Printf("  Default retention mode: %s\n", config.Rule.DefaultRetention.Mode)
	if config.Rule.DefaultRetention.Days > 0 {
		fmt.Printf("  Default retention days: %d\n", config.Rule.DefaultRetention.Days)
	}
	if config.Rule.DefaultRetention.Years > 0 {
		fmt.Printf("  Default retention years: %d\n", config.Rule.DefaultRetention.Years)
	}
}
