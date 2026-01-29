//go:build example
// +build example

// Example: List multipart uploads
package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "multipart-bucket"
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
	objectSvc := client.Object()

	result, err := objectSvc.ListMultipartUploads(ctx, YOURBUCKET,
		object.WithMultipartPrefix(""),
		object.WithMultipartMaxUploads(100),
	)
	if err != nil {
		log.Fatalf("Failed to list multipart uploads: %v", err)
	}

	if len(result.Uploads) == 0 {
		log.Printf("No multipart uploads found in %s", YOURBUCKET)
		return
	}

	log.Printf("Multipart uploads in %s:", YOURBUCKET)
	for _, upload := range result.Uploads {
		log.Printf("- Key: %s, UploadID: %s, Initiated: %s", upload.Key, upload.UploadID, upload.Initiated)
	}
}
