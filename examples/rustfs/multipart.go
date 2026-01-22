//go:build example

// multipart-new.go - Multipart upload example using new API
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/types"
)

func main() {
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "mybucket"
	)

	// Initialize client
	client, err := rustfs.New(YOURENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(YOURACCESSKEYID, YOURSECRETACCESSKEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalln("Failed to initialize client:", err)
	}

	ctx := context.Background()
	bucketName := YOURBUCKET
	objectName := "large-file.txt"

	// ===== Multipart upload using new API =====
	// Get Object service
	objectSvc := client.Object()

	// Type assertion to access multipart upload methods
	type MultipartService interface {
		InitiateMultipartUpload(ctx context.Context, bucketName, objectName string, opts ...object.PutOption) (string, error)
		UploadPart(ctx context.Context, bucketName, objectName, uploadID string, partNumber int, reader io.Reader, partSize int64, opts ...object.PutOption) (types.ObjectPart, error)
		CompleteMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string, parts []types.ObjectPart, opts ...object.PutOption) (types.UploadInfo, error)
		AbortMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string) error
	}

	multipartSvc, ok := objectSvc.(MultipartService)
	if !ok {
		log.Fatalln("Object service does not support multipart upload")
	}

	// 1. Initialize multipart upload
	log.Println("\n=== Initialize multipart upload ===")
	uploadID, err := multipartSvc.InitiateMultipartUpload(ctx, bucketName, objectName,
		object.WithContentType("text/plain"),
		object.WithUserMetadata(map[string]string{
			"upload-type": "multipart",
		}),
	)
	if err != nil {
		log.Fatalln("Failed to initialize multipart upload:", err)
	}
	log.Printf("✅ Initialization successful, Upload ID: %s\n", uploadID)

	// Defer abort (if error occurs)
	var uploadCompleted bool
	defer func() {
		if !uploadCompleted {
			log.Println("\n=== Abort multipart upload (cleanup) ===")
			err := multipartSvc.AbortMultipartUpload(ctx, bucketName, objectName, uploadID)
			if err != nil {
				log.Printf("Failed to abort multipart upload: %v\n", err)
			} else {
				log.Println("✅ Multipart upload aborted")
			}
		}
	}()

	// 2. Upload parts
	log.Println("\n=== Upload parts ===")
	parts := make([]types.ObjectPart, 0)

	// Simulate 3 parts (each part at least 5MB, last one can be less than 5MB)
	// Note: S3 requires each part (except the last one) to be at least 5MB
	partContents := []string{
		strings.Repeat("Part 1: This is the first part of the file. ", 120000),          // ~5.3MB
		strings.Repeat("Part 2: This is the second part of the file. ", 120000),         // ~5.4MB
		strings.Repeat("Part 3: This is the third and final part of the file. ", 50000), // ~2.5MB (last one can be less than 5MB)
	}

	for i, content := range partContents {
		partNumber := i + 1
		partData := strings.NewReader(content)
		partSize := int64(len(content))

		log.Printf("Uploading part %d/%d (size: %d bytes)...\n", partNumber, len(partContents), partSize)

		part, err := multipartSvc.UploadPart(ctx, bucketName, objectName, uploadID,
			partNumber, partData, partSize)
		if err != nil {
			log.Fatalf("Failed to upload part %d: %v\n", partNumber, err)
		}

		parts = append(parts, part)
		log.Printf("  ✅ Part %d uploaded successfully, ETag: %s\n", partNumber, part.ETag)
	}

	// 3. Complete multipart upload
	log.Println("\n=== Complete multipart upload ===")
	uploadInfo, err := multipartSvc.CompleteMultipartUpload(ctx, bucketName, objectName, uploadID, parts)
	if err != nil {
		log.Fatalln("Failed to complete multipart upload:", err)
	}

	uploadCompleted = true // Mark upload as completed to avoid abort
	log.Printf("✅ Multipart upload completed!\n")
	log.Printf("   Object: %s\n", uploadInfo.Key)
	log.Printf("   ETag: %s\n", uploadInfo.ETag)
	log.Printf("   Total size: %d bytes\n", uploadInfo.Size)

	// 4. Verify uploaded object
	log.Println("\n=== Verify uploaded object ===")
	objInfo, err := objectSvc.Stat(ctx, bucketName, objectName)
	if err != nil {
		log.Fatalln("Failed to get object info:", err)
	}
	log.Printf("Object info:\n")
	log.Printf("  Name: %s\n", objInfo.Key)
	log.Printf("  Size: %d bytes\n", objInfo.Size)
	log.Printf("  ETag: %s\n", objInfo.ETag)
	log.Printf("  Last modified: %s\n", objInfo.LastModified.Format("2006-01-02 15:04:05"))

	// 5. Download and display partial content
	log.Println("\n=== Download and display partial content ===")
	reader, _, err := objectSvc.Get(ctx, bucketName, objectName,
		object.WithGetRange(0, 99), // Download first 100 bytes
	)
	if err != nil {
		log.Fatalln("Failed to download object:", err)
	}
	defer reader.Close()

	buf := make([]byte, 100)
	n, _ := reader.Read(buf)
	log.Printf("First 100 bytes content:\n%s\n", string(buf[:n]))

	// 6. Cleanup (optional)
	// log.Println("\n=== Delete uploaded object ===")
	// err = objectSvc.Delete(ctx, bucketName, objectName)
	// if err != nil {
	// 	log.Printf("Failed to delete object: %v\n", err)
	// } else {
	// 	log.Printf("✅ Successfully deleted object: %s\n", objectName)
	// }

	log.Println("\n=== Multipart upload example completed ===")
	fmt.Println("\nTips:")
	fmt.Println("- Multipart upload is suitable for large files (>5MB)")
	fmt.Println("- Each part minimum 5MB (except the last part)")
	fmt.Println("- Maximum 10,000 parts supported")
	fmt.Println("- If upload fails, uploaded parts will be automatically cleaned up")
}
