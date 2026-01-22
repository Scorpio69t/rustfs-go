//go:build example
// +build example

// Example: Delete an object
// Demonstrates how to delete an object using the RustFS Go SDK
package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "mybucket"
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

	// Get Object service
	objectSvc := client.Object()

	// Object to delete
	bucketName := YOURBUCKET
	objectName := "object-to-delete.txt"

	// Delete the object
	err = objectSvc.Delete(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to delete object: %v", err)
	}

	log.Printf("✅ Deleted object: %s/%s", bucketName, objectName)
	log.Println("\nNote: Deleting a non-existent object does not return an error")
}
