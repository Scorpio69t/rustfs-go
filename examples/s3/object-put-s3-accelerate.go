//go:build example
// +build example

// Example: Upload object using S3 Accelerate
package main

import (
	"context"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "s3.amazonaws.com"
		YOURBUCKET          = "accelerate-bucket"
	)

	// Initialize RustFS client
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      true,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()
	objectSvc := client.Object()

	content := "accelerate upload example"
	_, err = objectSvc.Put(ctx, YOURBUCKET, "accelerate.txt", strings.NewReader(content), int64(len(content)),
		object.WithPutAccelerate(),
	)
	if err != nil {
		log.Fatalf("Failed to upload object with accelerate: %v", err)
	}

	log.Println("Uploaded object using S3 Accelerate.")
}
