//go:build example
// +build example

// objectops-new.go - Object operations example using new API
package main

import (
	"context"
	"io"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	const (
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "mybucket" // Use 'mc mb play/mybucket' to create bucket (if not exists)
	)

	// Initialize client
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false, // Set to true to use HTTPS
	})
	if err != nil {
		log.Fatalln("Failed to initialize client:", err)
	}

	ctx := context.Background()
	bucketName := YOURBUCKET
	objectName := "test-object.txt"

	// ===== Using new API =====
	// Get Object service
	objectSvc := client.Object()

	// 1. Upload object (from string)
	log.Println("\n=== Upload object (from string) ===")
	data := strings.NewReader("Hello, RustFS! This is a test object.")
	uploadInfo, err := objectSvc.Put(ctx, bucketName, objectName, data, int64(data.Len()),
		object.WithContentType("text/plain; charset=utf-8"),
		object.WithUserMetadata(map[string]string{
			"author":  "rustfs-go",
			"version": "1.0",
		}),
		object.WithUserTags(map[string]string{
			"category": "example",
			"env":      "development",
		}),
	)
	if err != nil {
		log.Fatalln("Failed to upload object:", err)
	}
	log.Printf("✅ Successfully uploaded object: %s\n", uploadInfo.Key)
	log.Printf("   ETag: %s\n", uploadInfo.ETag)
	log.Printf("   Size: %d bytes\n", uploadInfo.Size)
	if uploadInfo.VersionID != "" {
		log.Printf("   Version ID: %s\n", uploadInfo.VersionID)
	}

	// 2. Get object info
	log.Println("\n=== Get object info ===")
	objInfo, err := objectSvc.Stat(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalln("Failed to get object info:", err)
	}
	log.Printf("Object: %s\n", objInfo.Key)
	log.Printf("  Size: %d bytes\n", objInfo.Size)
	log.Printf("  Type: %s\n", objInfo.ContentType)
	log.Printf("  ETag: %s\n", objInfo.ETag)
	log.Printf("  Last modified: %s\n", objInfo.LastModified.Format("2006-01-02 15:04:05"))
	if len(objInfo.UserMetadata) > 0 {
		log.Println("  User metadata:")
		for k, v := range objInfo.UserMetadata {
			log.Printf("    %s: %s\n", k, v)
		}
	}

	// 3. Download object
	log.Println("\n=== Download object ===")
	reader, _, err := objectSvc.Get(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalln("Failed to download object:", err)
	}
	defer reader.Close()

	buf := make([]byte, 1024)
	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		log.Fatalln("Failed to read object content:", err)
	}
	log.Printf("Object content: %s\n", string(buf[:n]))

	// 4. Download part of object (Range request)
	log.Println("\n=== Download part of object (Range request) ===")
	rangeReader, _, err := objectSvc.Get(ctx, bucketName, objectName,
		object.WithGetRange(0, 10), // Download only first 11 bytes (0-10)
	)
	if err != nil {
		log.Fatalln("Range download failed:", err)
	}
	defer rangeReader.Close()

	rangeBuf := make([]byte, 20)
	n, _ = rangeReader.Read(rangeBuf)
	log.Printf("Partial content (0-10 bytes): %s\n", string(rangeBuf[:n]))

	// 5. List objects
	log.Printf("\n=== List objects in bucket %s ===\n", bucketName)
	objectsCh := objectSvc.List(ctx, bucketName)
	count := 0
	for obj := range objectsCh {
		if obj.Err != nil {
			log.Printf("Error listing objects: %v\n", obj.Err)
			break
		}
		count++
		log.Printf("  %d. %s (size: %d bytes)\n", count, obj.Key, obj.Size)
	}

	// 6. Copy object
	log.Println("\n=== Copy object ===")
	copyObjectName := "test-object-copy.txt"
	copyInfo, err := objectSvc.Copy(ctx,
		bucketName, copyObjectName, // Destination
		bucketName, objectName, // Source
		object.WithCopyMetadata(map[string]string{
			"copied": "true",
		}, true), // Replace metadata
	)
	if err != nil {
		log.Printf("Failed to copy object: %v\n", err)
	} else {
		log.Printf("✅ Successfully copied object: %s -> %s\n", objectName, copyObjectName)
		log.Printf("   New object ETag: %s\n", copyInfo.ETag)
	}

	// 7. Delete object
	// log.Println("\n=== Delete object ===")
	// err = objectSvc.Delete(ctx, bucketName, objectName)
	// if err != nil {
	// 	log.Fatalln("Failed to delete object:", err)
	// }
	// log.Printf("✅ Successfully deleted object: %s\n", objectName)

	// // Delete copied object
	// err = objectSvc.Delete(ctx, bucketName, copyObjectName)
	// if err != nil {
	// 	log.Printf("Failed to delete copied object: %v\n", err)
	// } else {
	// 	log.Printf("✅ Successfully deleted object: %s\n", copyObjectName)
	// }

	log.Println("\n=== Example completed ===")
}
