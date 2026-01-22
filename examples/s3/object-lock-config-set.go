//go:build example
// +build example

// Example: Set bucket object lock configuration
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/bucket"
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

	// Object Lock must be enabled when the bucket is created.
	exists, err := bucketSvc.Exists(ctx, YOURBUCKET)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := bucketSvc.Create(ctx, YOURBUCKET, bucket.WithObjectLocking(true)); err != nil {
			log.Fatalf("Failed to create bucket with object lock enabled: %v", err)
		}
	}

	config := objectlock.Config{
		ObjectLockEnabled: objectlock.ObjectLockEnabledValue,
		Rule: &objectlock.Rule{
			DefaultRetention: objectlock.DefaultRetention{
				Mode: objectlock.RetentionGovernance,
				Days: 7,
			},
		},
	}

	if err := bucketSvc.SetObjectLockConfig(ctx, YOURBUCKET, config); err != nil {
		log.Fatalf("Failed to set object lock configuration: %v", err)
	}

	fmt.Printf("Object lock configuration set for %s\n", YOURBUCKET)
}
