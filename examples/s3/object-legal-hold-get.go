//go:build example
// +build example

// Example: Get object legal hold status
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
		YOURBUCKET          = "object-lock-bucket"
	)

	const (
		OBJECTNAME      = "legal-hold-object.txt"
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

	var opts []object.LegalHoldOption
	if OBJECTVERSIONID != "" {
		opts = append(opts, object.WithLegalHoldVersionID(OBJECTVERSIONID))
	}

	status, err := objectSvc.GetLegalHold(ctx, YOURBUCKET, OBJECTNAME, opts...)
	if err != nil {
		log.Fatalf("Failed to get legal hold status: %v", err)
	}

	if OBJECTVERSIONID != "" {
		log.Printf("Legal hold status for %s/%s (version %s): %s", YOURBUCKET, OBJECTNAME, OBJECTVERSIONID, status)
		return
	}

	log.Printf("Legal hold status for %s/%s: %s", YOURBUCKET, OBJECTNAME, status)
}
