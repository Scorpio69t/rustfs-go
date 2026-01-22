//go:build example
// +build example

// Example: Listen for bucket notifications
package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	bucketSvc := client.Bucket()
	objectSvc := client.Object()

	exists, err := bucketSvc.Exists(ctx, YOURBUCKET)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := bucketSvc.Create(ctx, YOURBUCKET); err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
	}

	events := []notification.EventType{notification.ObjectCreatedAll}
	ch := bucketSvc.ListenNotification(ctx, YOURBUCKET, "", "", events)

	time.Sleep(500 * time.Millisecond)
	objectName := "notification-test.txt"
	content := "Notification test content."
	if _, err := objectSvc.Put(ctx, YOURBUCKET, objectName, strings.NewReader(content), int64(len(content)),
		object.WithContentType("text/plain"),
	); err != nil {
		log.Fatalf("Failed to upload test object: %v", err)
	}

	for {
		select {
		case info, ok := <-ch:
			if !ok {
				log.Println("Notification channel closed")
				return
			}
			if info.Err != nil {
				log.Fatalf("Notification error: %v", info.Err)
			}
			if len(info.Records) == 0 {
				continue
			}
			for _, record := range info.Records {
				log.Printf("Event: %s", record.EventName)
				log.Printf("Bucket: %s", record.S3.Bucket.Name)
				log.Printf("Object: %s", record.S3.Object.Key)
			}
			return
		case <-ctx.Done():
			log.Println("No notifications received before timeout")
			return
		}
	}
}
