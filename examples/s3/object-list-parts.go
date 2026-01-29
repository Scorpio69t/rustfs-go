//go:build example
// +build example

// Example: List multipart upload parts
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

	const (
		OBJECTNAME = "multipart-object.bin"
		UPLOADID   = "YOUR_UPLOAD_ID"
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

	result, err := objectSvc.ListObjectParts(ctx, YOURBUCKET, OBJECTNAME, UPLOADID, object.WithListPartsMax(100))
	if err != nil {
		log.Fatalf("Failed to list parts: %v", err)
	}

	if len(result.Parts) == 0 {
		log.Printf("No parts found for upload %s", UPLOADID)
		return
	}

	log.Printf("Parts for %s/%s (upload %s):", YOURBUCKET, OBJECTNAME, UPLOADID)
	for _, part := range result.Parts {
		log.Printf("- Part %d, Size %d, ETag %s", part.PartNumber, part.Size, part.ETag)
	}
}
