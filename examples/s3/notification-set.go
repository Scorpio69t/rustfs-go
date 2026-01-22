//go:build example
// +build example

// Example: Set bucket notification configuration
package main

import (
	"context"
	"log"

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

	exists, err := bucketSvc.Exists(ctx, YOURBUCKET)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := bucketSvc.Create(ctx, YOURBUCKET); err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
	}

	config := notification.Configuration{
		QueueConfigs: []notification.QueueConfig{
			{
				Config: notification.Config{
					ID:     "queue-notification",
					Events: []notification.EventType{notification.ObjectCreatedPut, notification.ObjectRemovedDelete},
					Filter: &notification.Filter{
						S3Key: notification.S3Key{
							FilterRules: []notification.FilterRule{
								{Name: "prefix", Value: "uploads/"},
							},
						},
					},
				},
				Queue: "arn:aws:sqs:us-east-1:123456789012:example-queue",
			},
		},
	}

	xmlData, err := config.ToXML()
	if err != nil {
		log.Fatalf("Failed to build notification configuration: %v", err)
	}

	if err := bucketSvc.SetNotification(ctx, YOURBUCKET, xmlData); err != nil {
		log.Fatalf("Failed to set notification configuration: %v", err)
	}

	log.Printf("Notification configuration set for %s", YOURBUCKET)
}
