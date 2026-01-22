//go:build example
// +build example

// object_tagging.go - Demonstrates fput/fget helpers, SSE, and object tagging
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	const (
		accessKey = "rustfsadmin"
		secretKey = "rustfsadmin"
		endpoint  = "127.0.0.1:9000"
		bucket    = "mybucket"
	)

	ctx := context.Background()

	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("failed to init client: %v", err)
	}

	obj := client.Object()

	// Prepare a small demo file for upload
	tmpFile, err := os.CreateTemp("", "rustfs-demo-*.txt")
	if err != nil {
		log.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString("hello from rustfs fput/fget example"); err != nil {
		log.Fatalf("failed to write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		log.Fatalf("failed to close temp file: %v", err)
	}

	// Upload using fput with user tags (enable SSE-S3 by adding object.WithSSES3() if your server has SSE configured)
	uploadInfo, err := obj.FPut(
		ctx,
		bucket,
		"demo/hello.txt",
		tmpFile.Name(),
		object.WithUserTags(map[string]string{"env": "dev", "team": "storage"}),
		object.WithContentType("text/plain"),
	)
	if err != nil {
		log.Fatalf("FPut failed: %v", err)
	}
	log.Printf("Uploaded object %s (etag=%s)", uploadInfo.Key, uploadInfo.ETag)

	// Read tags back
	tags, err := obj.GetTagging(ctx, bucket, uploadInfo.Key)
	if err != nil {
		log.Fatalf("GetTagging failed: %v", err)
	}
	log.Printf("Current tags: %+v", tags)

	// Download to a new path using fget
	targetPath := tmpFile.Name() + ".downloaded"
	downloadInfo, err := obj.FGet(ctx, bucket, uploadInfo.Key, targetPath)
	if err != nil {
		log.Fatalf("FGet failed: %v", err)
	}
	defer os.Remove(targetPath)

	log.Printf("Downloaded object %s to %s (size=%d)", downloadInfo.Key, targetPath, downloadInfo.Size)

	// Update tags then delete them
	if err := obj.SetTagging(ctx, bucket, uploadInfo.Key, map[string]string{"env": "prod"}); err != nil {
		log.Fatalf("SetTagging failed: %v", err)
	}
	if err := obj.DeleteTagging(ctx, bucket, uploadInfo.Key); err != nil {
		log.Fatalf("DeleteTagging failed: %v", err)
	}

	log.Printf("Tagging lifecycle done at %s", time.Now().Format(time.RFC3339))
}
