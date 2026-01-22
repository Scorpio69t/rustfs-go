//go:build example
// +build example

// Example: Set object ACL
// Demonstrates how to set a canned ACL on an object using the RustFS Go SDK.
package main

import (
	"context"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/acl"
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
	objectSvc := client.Object()

	bucketName := YOURBUCKET
	objectName := "acl-example-object.txt"
	content := "Hello from RustFS ACL example."

	// Ensure the object exists before setting ACL.
	_, err = objectSvc.Put(ctx, bucketName, objectName, strings.NewReader(content), int64(len(content)),
		object.WithContentType("text/plain; charset=utf-8"),
	)
	if err != nil {
		log.Fatalf("Failed to upload object: %v", err)
	}

	policy := acl.ACL{Canned: acl.ACLPublicRead}
	if err := objectSvc.SetACL(ctx, bucketName, objectName, policy); err != nil {
		log.Fatalf("Failed to set object ACL: %v", err)
	}

	log.Println("Object ACL set successfully.")
	log.Printf("Bucket: %s", bucketName)
	log.Printf("Object: %s", objectName)
	log.Printf("Canned ACL: %s", policy.Canned)
}
