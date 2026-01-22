//go:build example
// +build example

// Example: End-to-end scenario
// Demonstrates a full bucket lifecycle with configuration and object operations.
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/bucket"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/cors"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/sse"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
	)

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

	bucketName := fmt.Sprintf("e2e-%d", time.Now().Unix())
	objectName := "e2e-object.txt"
	content := "Hello, RustFS! End-to-end scenario validation."

	if err := bucketSvc.Create(ctx, bucketName, bucket.WithRegion("us-east-1")); err != nil {
		log.Fatalf("Failed to create bucket: %v", err)
	}
	defer func() {
		if err := bucketSvc.Delete(ctx, bucketName); err != nil {
			log.Printf("Cleanup bucket failed: %v", err)
		}
	}()

	corsConfig := cors.NewConfig([]cors.Rule{
		{
			AllowedOrigin: []string{"*"},
			AllowedMethod: []string{"GET", "PUT"},
			AllowedHeader: []string{"*"},
			MaxAgeSeconds: 3600,
		},
	})
	if err := bucketSvc.SetCORS(ctx, bucketName, corsConfig); err != nil {
		log.Fatalf("Failed to set CORS: %v", err)
	}
	if _, err := bucketSvc.GetCORS(ctx, bucketName); err != nil {
		log.Fatalf("Failed to get CORS: %v", err)
	}
	if err := bucketSvc.DeleteCORS(ctx, bucketName); err != nil {
		log.Fatalf("Failed to delete CORS: %v", err)
	}

	encConfig := sse.NewConfiguration()
	if err := bucketSvc.SetEncryption(ctx, bucketName, *encConfig); err != nil {
		log.Fatalf("Failed to set bucket encryption: %v", err)
	}
	if _, err := bucketSvc.GetEncryption(ctx, bucketName); err != nil {
		log.Fatalf("Failed to get bucket encryption: %v", err)
	}

	uploadInfo, err := objectSvc.Put(
		ctx,
		bucketName,
		objectName,
		strings.NewReader(content),
		int64(len(content)),
		object.WithContentType("text/plain; charset=utf-8"),
		object.WithSSES3(),
	)
	if err != nil {
		log.Fatalf("Failed to upload object: %v", err)
	}
	log.Printf("Uploaded object %s (etag=%s)", uploadInfo.Key, uploadInfo.ETag)

	statInfo, err := objectSvc.Stat(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to stat object: %v", err)
	}
	log.Printf("Stat object size=%d content-type=%s", statInfo.Size, statInfo.ContentType)

	reader, _, err := objectSvc.Get(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to get object: %v", err)
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		log.Fatalf("Failed to read object: %v", err)
	}
	if err := reader.Close(); err != nil {
		log.Fatalf("Failed to close object reader: %v", err)
	}
	if string(data) != content {
		log.Fatalf("Content mismatch: got %q", string(data))
	}

	tags := map[string]string{
		"scenario": "e2e",
		"owner":    "rustfs-go",
	}
	if err := objectSvc.SetTagging(ctx, bucketName, objectName, tags); err != nil {
		log.Fatalf("Failed to set tags: %v", err)
	}
	gotTags, err := objectSvc.GetTagging(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to get tags: %v", err)
	}
	log.Printf("Object tags: %+v", gotTags)

	if err := objectSvc.Delete(ctx, bucketName, objectName); err != nil {
		log.Fatalf("Failed to delete object: %v", err)
	}

	if err := bucketSvc.DeleteEncryption(ctx, bucketName); err != nil {
		log.Printf("Failed to delete bucket encryption: %v", err)
	}

	log.Println("End-to-end scenario completed successfully.")
}
