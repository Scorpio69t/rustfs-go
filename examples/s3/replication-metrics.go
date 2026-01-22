//go:build example
// +build example

// Example: Get bucket replication metrics
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

	metrics, err := bucketSvc.GetReplicationMetrics(ctx, SOURCEBUCKET)
	if err != nil {
		log.Fatalf("Failed to get replication metrics: %v", err)
	}

	log.Printf("Replication metrics for %s", SOURCEBUCKET)
	log.Printf("Replicated count: %d", metrics.ReplicatedCount)
	log.Printf("Replicated size: %d", metrics.ReplicatedSize)
	log.Printf("Replica count: %d", metrics.ReplicaCount)
	log.Printf("Replica size: %d", metrics.ReplicaSize)
	log.Printf("Pending count: %d", metrics.PendingCount)
	log.Printf("Pending size: %d", metrics.PendingSize)
	log.Printf("Failed count: %d", metrics.FailedCount)
	log.Printf("Failed size: %d", metrics.FailedSize)

	if len(metrics.Stats) == 0 {
		log.Println("No per-target metrics available")
		return
	}

	for target, stat := range metrics.Stats {
		log.Printf("Target: %s", target)
		log.Printf("  Replicated count: %d", stat.ReplicatedCount)
		log.Printf("  Replicated size: %d", stat.ReplicatedSize)
		log.Printf("  Pending count: %d", stat.PendingCount)
		log.Printf("  Pending size: %d", stat.PendingSize)
		log.Printf("  Failed count: %d", stat.FailedCount)
		log.Printf("  Failed size: %d", stat.FailedSize)
	}
}
