//go:build example
// +build example

// Example: Get bucket replication configuration
package main

import (
	"context"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/replication"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
		SOURCEBUCKET        = "replication-source"
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

	data, err := bucketSvc.GetReplication(ctx, SOURCEBUCKET)
	if err != nil {
		log.Fatalf("Failed to get replication configuration: %v", err)
	}

	config, err := replication.ParseConfig(strings.NewReader(string(data)))
	if err != nil {
		log.Fatalf("Failed to parse replication configuration: %v", err)
	}

	log.Printf("Replication configuration for %s", SOURCEBUCKET)
	log.Printf("Role: %s", config.Role)
	log.Printf("Rules: %d", len(config.Rules))

	for _, rule := range config.Rules {
		log.Printf("Rule ID: %s", rule.ID)
		log.Printf("  Status: %s", rule.Status)
		log.Printf("  Priority: %d", rule.Priority)
		log.Printf("  Filter prefix: %s", rule.Filter.Prefix)
		log.Printf("  Destination: %s", rule.Destination.Bucket)
	}
}
