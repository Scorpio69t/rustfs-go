//go:build example
// +build example

// Example: Upload object with checksum mode
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
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "checksum-bucket"
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

	content := "checksum mode example"
	_, err = objectSvc.Put(ctx, YOURBUCKET, "checksum.txt", strings.NewReader(content), int64(len(content)),
		object.WithChecksumMode("ENABLED"),
		object.WithChecksumAlgorithm("CRC32C"),
	)
	if err != nil {
		log.Fatalf("Failed to upload object: %v", err)
	}

	log.Println("Uploaded object with checksum mode enabled.")
}
