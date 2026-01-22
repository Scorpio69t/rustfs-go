//go:build example
// +build example

// Example: Restore an archived object
// Demonstrates how to submit a restore request for objects stored in archival storage.
package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/restore"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "archive-bucket"
	)

	const (
		OBJECTNAME      = "archived-object.txt"
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

	restoreReq := restore.RestoreRequest{}
	restoreReq.SetDays(7)
	restoreReq.SetTier(restore.TierBulk)

	if err := objectSvc.Restore(ctx, YOURBUCKET, OBJECTNAME, OBJECTVERSIONID, restoreReq); err != nil {
		log.Fatalf("Failed to submit restore request: %v", err)
	}

	if OBJECTVERSIONID != "" {
		log.Printf("Restore request submitted for %s/%s (version %s).", YOURBUCKET, OBJECTNAME, OBJECTVERSIONID)
	} else {
		log.Printf("Restore request submitted for %s/%s.", YOURBUCKET, OBJECTNAME)
	}
	log.Println("Restored objects may take time to become available.")
}
