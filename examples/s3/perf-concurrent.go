//go:build example
// +build example

// Example: Concurrent operations performance
// Runs parallel uploads to measure throughput under concurrency.
package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/bucket"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
	)

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

	bucketName := fmt.Sprintf("perf-concurrent-%d", time.Now().Unix())
	if err := bucketSvc.Create(ctx, bucketName, bucket.WithRegion("us-east-1")); err != nil {
		log.Fatalf("Failed to create bucket: %v", err)
	}
	defer func() {
		if err := bucketSvc.Delete(ctx, bucketName); err != nil {
			log.Printf("Cleanup bucket failed: %v", err)
		}
	}()

	workers := 8
	objectsPerWorker := 5
	payload := strings.Repeat("P", 256*1024) // 256 KB

	var wg sync.WaitGroup
	errCh := make(chan error, workers*objectsPerWorker)

	start := time.Now()
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for i := 0; i < objectsPerWorker; i++ {
				objectName := fmt.Sprintf("concurrent-%d-%d.bin", workerID, i)
				if _, err := objectSvc.Put(
					ctx,
					bucketName,
					objectName,
					strings.NewReader(payload),
					int64(len(payload)),
					object.WithContentType("application/octet-stream"),
				); err != nil {
					errCh <- fmt.Errorf("upload %s: %w", objectName, err)
					return
				}
				if err := objectSvc.Delete(ctx, bucketName, objectName); err != nil {
					errCh <- fmt.Errorf("delete %s: %w", objectName, err)
					return
				}
			}
		}(w)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			log.Fatalf("Concurrent run failed: %v", err)
		}
	}

	elapsed := time.Since(start)
	totalObjects := workers * objectsPerWorker
	log.Printf("Concurrent uploads: %d objects in %s", totalObjects, elapsed)
	log.Println("Concurrent performance test completed.")
}
