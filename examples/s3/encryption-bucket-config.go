// Example: Configure bucket default encryption
//
// When a bucket default encryption configuration is set, all new objects
// uploaded to the bucket will be encrypted by default.
package main

import (
	"context"
	"fmt"
	"log"

	rustfs "github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/sse"
)

func main() {
	// Connection configuration
	const (
		ACCESS_KEY = "XhJOoEKn3BM6cjD2dVmx"
		SECRET_KEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		ENDPOINT   = "127.0.0.1:9000"
	)

	bucketName := "encrypted-bucket"

	// Create RustFS client
	client, err := rustfs.New(ENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(ACCESS_KEY, SECRET_KEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("Failed to initialize client: %v", err)
	}

	ctx := context.Background()
	bucketSvc := client.Bucket()

	// Create bucket if absent
	exists, err := bucketSvc.Exists(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		err = bucketSvc.Create(ctx, bucketName)
		if err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
		fmt.Printf("✓ Bucket created: %s\n", bucketName)
	}

	// Create SSE-S3 configuration
	encryptionConfig := sse.NewConfiguration()

	fmt.Printf("\nSetting bucket default encryption...\n")
	fmt.Printf("  Algorithm: %s\n", encryptionConfig.Rules[0].ApplySSEByDefault.SSEAlgorithm)

	// Apply bucket encryption
	err = bucketSvc.SetEncryption(ctx, bucketName, *encryptionConfig)
	if err != nil {
		log.Fatalf("Failed to set bucket encryption: %v", err)
	}
	fmt.Printf("✓ Bucket default encryption set\n")

	// Retrieve bucket encryption
	fmt.Printf("\nRetrieving bucket encryption...\n")
	retrievedConfig, err := bucketSvc.GetEncryption(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to get bucket encryption: %v", err)
	}

	fmt.Printf("✓ Bucket encryption configuration:\n")
	for i, rule := range retrievedConfig.Rules {
		fmt.Printf("  Rule %d:\n", i+1)
		fmt.Printf("    Algorithm: %s\n", rule.ApplySSEByDefault.SSEAlgorithm)
		fmt.Printf("    BucketKeyEnabled: %v\n", rule.BucketKeyEnabled)
		if rule.ApplySSEByDefault.KMSMasterKeyID != "" {
			fmt.Printf("    KMS Key ID: %s\n", rule.ApplySSEByDefault.KMSMasterKeyID)
		}
	}

	// Demonstrate SSE-KMS configuration (optional)
	fmt.Printf("\nDemo: SSE-KMS configuration\n")
	kmsKeyID := "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
	kmsConfig := sse.NewKMSConfiguration(kmsKeyID)

	fmt.Printf("  KMS Key ID: %s\n", kmsConfig.Rules[0].ApplySSEByDefault.KMSMasterKeyID)
	fmt.Printf("  Algorithm: %s\n", kmsConfig.Rules[0].ApplySSEByDefault.SSEAlgorithm)

	// Note: Setting KMS requires a valid KMS key in a real environment
	// err = bucketSvc.SetEncryption(ctx, bucketName, *kmsConfig)
	// if err != nil {
	//  log.Printf("Warning: failed to set KMS encryption (requires valid KMS key): %v", err)
	// }

	// Delete encryption configuration
	fmt.Printf("\nDeleting bucket encryption configuration...\n")
	err = bucketSvc.DeleteEncryption(ctx, bucketName)
	if err != nil {
		log.Fatalf("Failed to delete bucket encryption: %v", err)
	}
	fmt.Printf("✓ Bucket encryption deleted\n")

	// Verify deletion
	fmt.Printf("\nVerifying deletion...\n")
	_, err = bucketSvc.GetEncryption(ctx, bucketName)
	if err != nil {
		if err == sse.ErrNoEncryptionConfig {
			fmt.Printf("✓ Confirmed: no encryption configuration for bucket\n")
		} else {
			log.Printf("Error retrieving encryption config: %v", err)
		}
	} else {
		fmt.Printf("⚠️  Warning: encryption config still present after delete\n")
	}

	fmt.Printf("\nBucket encryption notes:\n")
	fmt.Printf("  ✓ Default encryption applies to new objects\n")
	fmt.Printf("  ✓ Existing objects are not affected\n")
	fmt.Printf("  ✓ Supports SSE-S3 and SSE-KMS\n")
	fmt.Printf("  ✓ SSE-C cannot be used as bucket default encryption\n")
	fmt.Printf("  ✓ Enable default encryption for buckets containing sensitive data\n")

	// Cleanup (optional)
	// err = bucketSvc.Delete(ctx, bucketName)
	// if err != nil {
	//  log.Printf("Warning: failed to delete bucket: %v", err)
	// }
}
