//go:build example
// +build example

// Example: List objects with max-keys and pagination (StartAfter)
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
		YOURBUCKET          = "list-bucket"
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

	startAfter := "prefix/after-this-object.txt"
	startAfterOption := func(opts *object.ListOptions) {
		opts.StartAfter = startAfter
	}

	log.Printf("Listing objects after %q (max 5 keys)", startAfter)
	for info := range objectSvc.List(ctx, YOURBUCKET,
		object.WithListMaxKeys(5),
		startAfterOption,
	) {
		if info.Err != nil {
			log.Fatalf("List error: %v", info.Err)
		}
		log.Printf("Object: %s (size=%d)", info.Key, info.Size)
	}
}
