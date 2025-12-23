//go:build example
// +build example

// presigned.go - Presigned URL examples using the new Object service API
package main

import (
	"context"
	"log"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
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

	// Generate presigned GET URL (valid for 15 minutes) with a response header override
	getURL, getSignedHeaders, err := client.Object().PresignGet(
		ctx,
		bucketName,
		objectName,
		15*time.Minute,
		url.Values{"response-content-type": []string{"text/plain"}},
	)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Presigned GET URL (15m): %s", getURL.String())
	if len(getSignedHeaders) > 0 {
		log.Printf("Signed headers for GET: %+v", getSignedHeaders)
	}

	// Generate presigned PUT URL (valid for 15 minutes) signing SSE-S3 header
	putURL, putSignedHeaders, err := client.Object().PresignPut(
		ctx,
		bucketName,
		"upload-object.txt",
		15*time.Minute,
		nil,
		object.WithPresignSSES3(),
	)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Presigned PUT URL (15m, SSE-S3): %s", putURL.String())
	if len(putSignedHeaders) > 0 {
		log.Printf("Signed headers for PUT: %+v", putSignedHeaders)
	}
}
