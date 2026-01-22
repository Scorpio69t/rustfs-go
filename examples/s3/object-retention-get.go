//go:build example
// +build example

// Example: Get object retention
package main

import (
	"context"
	"log"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "object-lock-bucket"
	)

	const (
		OBJECTNAME      = "retention-object.txt"
		OBJECTVERSIONID = ""
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

	var opts []object.RetentionOption
	if OBJECTVERSIONID != "" {
		opts = append(opts, object.WithRetentionVersionID(OBJECTVERSIONID))
	}

	mode, retainUntil, err := objectSvc.GetRetention(ctx, YOURBUCKET, OBJECTNAME, opts...)
	if err != nil {
		log.Fatalf("Failed to get retention: %v", err)
	}

	log.Printf("Retention for %s/%s:", YOURBUCKET, OBJECTNAME)
	if OBJECTVERSIONID != "" {
		log.Printf("VersionID: %s", OBJECTVERSIONID)
	}

	if mode == "" && retainUntil.IsZero() {
		log.Println("No retention configuration found")
		return
	}

	log.Printf("Mode: %s", mode)
	if !retainUntil.IsZero() {
		log.Printf("Retain until: %s", retainUntil.UTC().Format(time.RFC3339))
	}
}
