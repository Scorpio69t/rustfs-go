//go:build example
// +build example

// Example: Get object ACL
// Demonstrates how to retrieve ACL grants for an object using the RustFS Go SDK.
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
	objectSvc := client.Object()

	bucketName := YOURBUCKET
	objectName := "acl-example-object.txt"

	policy, err := objectSvc.GetACL(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalf("Failed to get object ACL: %v", err)
	}

	log.Println("Object ACL retrieved successfully.")
	log.Printf("Bucket: %s", bucketName)
	log.Printf("Object: %s", objectName)
	log.Printf("Owner ID: %s", policy.Owner.ID)
	log.Printf("Owner DisplayName: %s", policy.Owner.DisplayName)
	log.Printf("Grant count: %d", len(policy.Grants))
	for i, grant := range policy.Grants {
		grantee := grant.Grantee
		log.Printf("Grant %d: %s %s %s (%s)",
			i+1,
			grant.Permission,
			grantee.Type,
			grantee.ID,
			grantee.URI,
		)
	}
}
