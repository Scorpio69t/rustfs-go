//go:build example
// +build example

// Example: Set bucket replication configuration
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/replication"
	"github.com/Scorpio69t/rustfs-go/types"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
		SOURCEBUCKET        = "replication-source"
		DESTBUCKET          = "replication-dest"
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

	for _, name := range []string{SOURCEBUCKET, DESTBUCKET} {
		exists, err := bucketSvc.Exists(ctx, name)
		if err != nil {
			log.Fatalf("Failed to check bucket %s: %v", name, err)
		}
		if !exists {
			if err := bucketSvc.Create(ctx, name); err != nil {
				log.Fatalf("Failed to create bucket %s: %v", name, err)
			}
		}
		if err := bucketSvc.SetVersioning(ctx, name, types.VersioningConfig{Status: "Enabled"}); err != nil {
			log.Fatalf("Failed to enable versioning on %s: %v", name, err)
		}
	}

	destArn := fmt.Sprintf("arn:aws:s3:::%s", DESTBUCKET)
	config := replication.Config{
		Role: "arn:aws:iam::123456789012:role/replication",
		Rules: []replication.Rule{
			{
				ID:       "example-replication-rule",
				Status:   replication.Enabled,
				Priority: 1,
				Filter: replication.Filter{
					Prefix: "logs/",
				},
				Destination: replication.Destination{
					Bucket:       destArn,
					StorageClass: "STANDARD",
				},
			},
		},
	}

	xmlData, err := config.ToXML()
	if err != nil {
		log.Fatalf("Failed to build replication configuration: %v", err)
	}

	if err := bucketSvc.SetReplication(ctx, SOURCEBUCKET, xmlData); err != nil {
		log.Fatalf("Failed to set replication configuration: %v", err)
	}

	log.Printf("Replication configuration set for %s -> %s", SOURCEBUCKET, destArn)
}
