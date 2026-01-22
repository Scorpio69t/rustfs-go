//go:build example
// +build example

// Example: Set object retention
package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/bucket"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/objectlock"
	"github.com/Scorpio69t/rustfs-go/types"
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
	objectSvc := client.Object()

	exists, err := bucketSvc.Exists(ctx, YOURBUCKET)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := bucketSvc.Create(ctx, YOURBUCKET, bucket.WithObjectLocking(true)); err != nil {
			log.Fatalf("Failed to create bucket with object lock enabled: %v", err)
		}
	}

	if err := bucketSvc.SetVersioning(ctx, YOURBUCKET, types.VersioningConfig{Status: "Enabled"}); err != nil {
		log.Fatalf("Failed to enable versioning: %v", err)
	}

	objectName := "retention-object.txt"
	content := "Object retention example."
	reader := strings.NewReader(content)

	uploadInfo, err := objectSvc.Put(ctx, YOURBUCKET, objectName, reader, int64(len(content)))
	if err != nil {
		log.Fatalf("Failed to upload object: %v", err)
	}

	retainUntil := time.Now().UTC().Add(24 * time.Hour)

	var opts []object.RetentionOption
	if uploadInfo.VersionID != "" {
		opts = append(opts, object.WithRetentionVersionID(uploadInfo.VersionID))
	}

	if err := objectSvc.SetRetention(ctx, YOURBUCKET, objectName, objectlock.RetentionGovernance, retainUntil, opts...); err != nil {
		log.Fatalf("Failed to set retention: %v", err)
	}

	log.Printf("Retention set for %s/%s until %s", YOURBUCKET, objectName, retainUntil.Format(time.RFC3339))
	if uploadInfo.VersionID != "" {
		log.Printf("VersionID: %s", uploadInfo.VersionID)
	}
}
