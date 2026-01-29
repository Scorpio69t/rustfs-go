//go:build example
// +build example

// Example: Copy object and replace tags
package main

import (
	"context"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	obj "github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "copy-bucket"
	)

	const (
		SOURCEOBJECT = "source-object.txt"
		DESTOBJECT   = "copied-object.txt"
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
	bucketSvc := client.Bucket()
	objectSvc := client.Object()

	exists, err := bucketSvc.Exists(ctx, YOURBUCKET)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := bucketSvc.Create(ctx, YOURBUCKET); err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
	}

	content := "Copy with new tags example."
	if _, err := objectSvc.Put(ctx, YOURBUCKET, SOURCEOBJECT, strings.NewReader(content), int64(len(content))); err != nil {
		log.Fatalf("Failed to upload source object: %v", err)
	}

	newTags := map[string]string{
		"env":  "test",
		"team": "storage",
	}

	_, err = objectSvc.Copy(ctx, YOURBUCKET, DESTOBJECT, YOURBUCKET, SOURCEOBJECT, func(opts *obj.CopyOptions) {
		opts.ReplaceTagging = true
		opts.UserTags = newTags
	})
	if err != nil {
		log.Fatalf("Failed to copy object with new tags: %v", err)
	}

	log.Printf("Copied %s to %s with new tags", SOURCEOBJECT, DESTOBJECT)
}
