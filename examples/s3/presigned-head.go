//go:build example
// +build example

// Example: Generate a presigned HEAD URL
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
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "presign-bucket"
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

	reqParams := url.Values{
		"response-content-disposition": []string{"inline"},
	}

	presignedURL, headers, err := objectSvc.PresignHead(ctx, YOURBUCKET, "demo.txt", 15*time.Minute, reqParams)
	if err != nil {
		log.Fatalf("Failed to presign HEAD URL: %v", err)
	}

	log.Println("✅ Presigned HEAD URL generated successfully!")
	log.Println("\nPresigned URL:")
	log.Println(presignedURL.String())

	if len(headers) > 0 {
		log.Println("\nSigned headers:")
		for k, v := range headers {
			log.Printf("%s: %s", k, v)
		}
	}
}
