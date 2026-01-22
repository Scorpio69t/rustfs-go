//go:build example

// bucketops-new.go - using new bucket API to perform bucket operations
package main

import (
	"context"
	"log"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/bucket"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "mybucket"
	)

	// create RustFS client
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false, // set to true if using HTTPS
	})
	if err != nil {
		log.Fatalln("Initialize RustFS client failed:", err)
	}

	ctx := context.Background()
	bucketName := YOURBUCKET

	// get Bucket service
	bucketSvc := client.Bucket()

	// 1. create bucket
	log.Println("\n=== Create Bucket ===")
	err = bucketSvc.Create(ctx, bucketName,
		bucket.WithRegion("us-east-1"),
		// bucket.WithObjectLocking(true), // optional: enable object locking
	)
	if err != nil {
		log.Printf("Create bucket failed: %v\n", err)
	} else {
		log.Printf("✅ Successfully created bucket: %s\n", bucketName)
	}

	// 2. check if bucket exists
	log.Println("\n=== Check Bucket Existence ===")
	exists, err := bucketSvc.Exists(ctx, bucketName)
	if err != nil {
		log.Fatalln("Check bucket existence failed:", err)
	}
	log.Printf("Bucket %s exists: %t\n", bucketName, exists)

	// 3. get bucket location
	log.Println("\n=== Get Bucket Location ===")
	location, err := bucketSvc.GetLocation(ctx, bucketName)
	if err != nil {
		log.Fatalln("Get bucket location failed:", err)
	}
	log.Printf("Bucket %s is located in region: %s\n", bucketName, location)

	// 4. list all buckets
	log.Println("\n=== List All Buckets ===")
	buckets, err := bucketSvc.List(ctx)
	if err != nil {
		log.Fatalln("List buckets failed:", err)
	}
	log.Printf("Found %d buckets:\n", len(buckets))
	for i, b := range buckets {
		log.Printf("  %d. %s (Create at: %s)\n",
			i+1, b.Name, b.CreationDate.Format("2006-01-02 15:04:05"))
	}

	// 5. list objects in the bucket
	log.Printf("\n=== List Objects in Bucket: %s ===\n", bucketName)
	objectSvc := client.Object()
	objectsCh := objectSvc.List(ctx, bucketName)
	count := 0
	for obj := range objectsCh {
		if obj.Err != nil {
			log.Printf("List objects failed: %v\n", obj.Err)
			break
		}
		count++
		if obj.IsPrefix {
			log.Printf("  %d. %s (Folder)\n", count, obj.Key)
			continue
		}
		log.Printf("  %d. %s (Size: %d bytes, Modified: %s)\n",
			count, obj.Key, obj.Size, obj.LastModified.Format("2006-01-02 15:04:05"))
	}
	if count == 0 {
		log.Println("  Bucket is empty.")
	}

	// 6. delete bucket
	// log.Println("\n=== Delete Bucket ===")
	// err = bucketSvc.Delete(ctx, bucketName)

	// if err != nil {
	// 	log.Printf("Delete bucket failed: %v\n", err)
	// } else {
	// 	log.Printf("✅ Successfully deleted bucket: %s\n", bucketName)
	// }

	log.Println("\n=== Bucket Operations Completed ===")
}
