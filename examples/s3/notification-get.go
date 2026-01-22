//go:build example
// +build example

// Example: Get bucket notification configuration
package main

import (
	"context"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/notification"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "notification-bucket"
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

	data, err := bucketSvc.GetNotification(ctx, YOURBUCKET)
	if err != nil {
		log.Fatalf("Failed to get notification configuration: %v", err)
	}

	config, err := notification.ParseConfig(strings.NewReader(string(data)))
	if err != nil {
		log.Fatalf("Failed to parse notification configuration: %v", err)
	}

	log.Printf("Notification configuration for %s", YOURBUCKET)
	log.Printf("Queue targets: %d", len(config.QueueConfigs))
	log.Printf("Topic targets: %d", len(config.TopicConfigs))
	log.Printf("Lambda targets: %d", len(config.LambdaConfigs))

	for _, q := range config.QueueConfigs {
		log.Printf("Queue ID: %s", q.ID)
		log.Printf("  Queue ARN: %s", q.Queue)
		log.Printf("  Events: %v", q.Events)
	}
}
