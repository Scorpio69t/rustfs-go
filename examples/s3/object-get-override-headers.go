//go:build example
// +build example

// Example: Get object with response header overrides
package main

import (
	"context"
	"log"
	"net/url"

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
		YOURBUCKET          = "demo-bucket"
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

	overrides := url.Values{
		"response-content-type":        []string{"text/plain"},
		"response-content-disposition": []string{"inline"},
	}

	reader, _, err := objectSvc.Get(ctx, YOURBUCKET, "demo.txt", object.WithGetResponseHeaders(overrides))
	if err != nil {
		log.Fatalf("Failed to get object: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			log.Fatalf("Failed to close reader: %v", err)
		}
	}()

	log.Println("Object downloaded with response header overrides.")
}
