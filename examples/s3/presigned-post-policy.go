//go:build example
// +build example

// Example: Create a presigned POST policy for browser uploads
// Demonstrates how to generate a POST URL and form fields.
package main

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/policy"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
	)

	bucketName := "post-policy-bucket"
	objectPrefix := "uploads/"

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

	// Ensure bucket exists
	exists, err := bucketSvc.Exists(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := bucketSvc.Create(ctx, bucketName); err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
		fmt.Printf("Bucket created: %s\n", bucketName)
	}

	postPolicy := policy.NewPostPolicy()
	if err := postPolicy.SetBucket(bucketName); err != nil {
		log.Fatalf("Failed to set bucket: %v", err)
	}
	if err := postPolicy.SetKeyStartsWith(objectPrefix); err != nil {
		log.Fatalf("Failed to set key prefix: %v", err)
	}
	if err := postPolicy.SetContentLengthRange(1, 10*1024*1024); err != nil {
		log.Fatalf("Failed to set content length range: %v", err)
	}
	if err := postPolicy.SetExpires(time.Now().UTC().Add(10 * time.Minute)); err != nil {
		log.Fatalf("Failed to set expiration: %v", err)
	}

	postURL, formData, err := client.Object().PresignedPostPolicy(ctx, postPolicy)
	if err != nil {
		log.Fatalf("Failed to create presigned POST policy: %v", err)
	}

	// Allow the browser to inject the filename.
	formData["key"] = objectPrefix + "${filename}"

	fmt.Printf("POST URL: %s\n", postURL.String())
	fmt.Println("Form fields:")

	keys := make([]string, 0, len(formData))
	for k := range formData {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("  %s: %s\n", key, formData[key])
	}

	fmt.Println("\nNext step: open browser-upload.html and paste the URL and fields.")
}
