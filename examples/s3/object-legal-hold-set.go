//go:build example
// +build example

// Example: Set object legal hold
package main

import (
	"context"
	"log"
	"strings"

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

	objectName := "legal-hold-object.txt"
	content := "Object legal hold example."
	reader := strings.NewReader(content)

	uploadInfo, err := objectSvc.Put(ctx, YOURBUCKET, objectName, reader, int64(len(content)))
	if err != nil {
		log.Fatalf("Failed to upload object: %v", err)
	}

	var opts []object.LegalHoldOption
	if uploadInfo.VersionID != "" {
		opts = append(opts, object.WithLegalHoldVersionID(uploadInfo.VersionID))
	}

	if err := objectSvc.SetLegalHold(ctx, YOURBUCKET, objectName, objectlock.LegalHoldOn, opts...); err != nil {
		log.Fatalf("Failed to set legal hold: %v", err)
	}

	log.Printf("Legal hold enabled for %s/%s", YOURBUCKET, objectName)
	if uploadInfo.VersionID != "" {
		log.Printf("VersionID: %s", uploadInfo.VersionID)
	}
}
