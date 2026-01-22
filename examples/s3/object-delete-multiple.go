//go:build example
// +build example

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

const (
	endpoint  = "127.0.0.1:9000"
	accessKey = "rustfsadmin"
	secretKey = "rustfsadmin"
	bucket    = "mybucket"
)

func main() {
	// Create client
	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
	})
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	service := client.Object()

	// Define objects to delete
	objectsToDelete := []string{
		"test-object-1.txt",
		"test-object-2.txt",
		"test-object-3.txt",
	}

	fmt.Printf("Preparing to delete %d objects...\n", len(objectsToDelete))

	// Delete objects in a loop (API currently demonstrates per-object delete)
	deletedCount := 0
	for _, objectName := range objectsToDelete {
		err := service.Delete(ctx, bucket, objectName)
		if err != nil {
			fmt.Printf("Warning: failed to delete '%s': %v\n", objectName, err)
			continue
		}
		deletedCount++
		fmt.Printf("✅ Deleted: %s\n", objectName)
	}

	fmt.Printf("\nDone: %d succeeded, %d failed\n", deletedCount, len(objectsToDelete)-deletedCount)
}
