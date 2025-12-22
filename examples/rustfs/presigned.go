//go:build example
// +build example

// presigned.go - Presigned URL example (using old API)
// Note: Presigned URL functionality has not been migrated to new API yet, this example still uses old API
package main

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "mybucket" // 'mc mb play/mybucket' if it does not exist.
	)

	// Initialize client
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	bucketName := YOURBUCKET
	objectName := "test-object.txt"

	// Generate presigned GET URL (valid for 1 hour)
	presignedURL, err := client.PresignedGetObject(ctx, bucketName, objectName, time.Hour, url.Values{})
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Presigned GET URL (valid for 1 hour):\n%s\n", presignedURL.String())

	// Generate presigned PUT URL (valid for 1 hour)
	presignedPutURL, err := client.PresignedPutObject(ctx, bucketName, "upload-object.txt", time.Hour)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Presigned PUT URL (valid for 1 hour):\n%s\n", presignedPutURL.String())

	// Generate presigned POST URL
	policy := rustfs.NewPostPolicy()
	err = policy.SetExpires(time.Now().Add(time.Hour))
	if err != nil {
		log.Fatalln(err)
		return
	}

	err = policy.SetCondition("$eq", "bucket", bucketName)
	if err != nil {
		log.Fatalln(err)
		return
	}

	err = policy.SetCondition("$eq", "key", "post-object.txt")
	if err != nil {
		log.Fatalln(err)
		return
	}

	err = policy.SetCondition("$eq", "Content-Type", "text/plain")
	if err != nil {
		log.Fatalln(err)
		return
	}

	postURL, formData, err := client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Presigned POST URL:\n%s\n", postURL.String())
	log.Println("Form data:")
	for k, v := range formData {
		log.Printf("  %s: %s\n", k, v)
	}
}
