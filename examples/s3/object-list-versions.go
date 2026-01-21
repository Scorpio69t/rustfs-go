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
	accessKey = "XhJOoEKn3BM6cjD2dVmx"
	secretKey = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
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

	fmt.Printf("Listing all object versions in bucket '%s'...\n\n", bucket)

	// List object versions using ListVersions
	objectCh := service.ListVersions(ctx, bucket)

	versionCount := 0
	currentCount := 0

	for obj := range objectCh {
		if obj.Err != nil {
			fmt.Printf("Error: %v\n", obj.Err)
			continue
		}

		if obj.IsLatest {
			currentCount++
			fmt.Printf("📄 Object: %s\n", obj.Key)
			fmt.Printf("   VersionID: %s (current)\n", obj.VersionID)
		} else {
			versionCount++
			fmt.Printf("📋 Object: %s\n", obj.Key)
			fmt.Printf("   VersionID: %s\n", obj.VersionID)
		}

		fmt.Printf("   Size: %d bytes\n", obj.Size)
		fmt.Printf("   LastModified: %s\n", obj.LastModified.Format("2006-01-02 15:04:05"))
		if obj.IsDeleteMarker {
			fmt.Printf("   ⚠️  Delete marker\n")
		}
		fmt.Println()
	}

	fmt.Printf("Total: %d current versions, %d historical versions\n", currentCount, versionCount)
}
