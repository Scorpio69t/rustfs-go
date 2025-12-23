//go:build example
// +build example

// bucket_policy_lifecycle.go - Demonstrates bucket policy and lifecycle operations
package main

import (
	"context"
	"log"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func main() {
	const (
		accessKey  = "XhJOoEKn3BM6cjD2dVmx"
		secretKey  = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
		endpoint   = "127.0.0.1:9000"
		bucketName = "mybucket"
	)

	client, err := rustfs.New(endpoint, &rustfs.Options{
		Credentials: credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("failed to init client: %v", err)
	}

	bkt := client.Bucket()
	ctx := context.Background()

	// Set a simple bucket policy (public read on a prefix)
	policy := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::` + bucketName + `/public/*"]
    }
  ]
}`

	if err := bkt.SetPolicy(ctx, bucketName, policy); err != nil {
		log.Fatalf("SetPolicy failed: %v", err)
	}
	gotPolicy, err := bkt.GetPolicy(ctx, bucketName)
	if err != nil {
		log.Fatalf("GetPolicy failed: %v", err)
	}
	log.Printf("Bucket policy set. Length=%d bytes", len(gotPolicy))

	// Configure a simple lifecycle rule to expire objects after 30 days
	lifecycle := []byte(`<LifecycleConfiguration>
  <Rule>
    <ID>expire-temp</ID>
    <Status>Enabled</Status>
    <Filter><Prefix>temp/</Prefix></Filter>
    <Expiration><Days>30</Days></Expiration>
  </Rule>
</LifecycleConfiguration>`)

	if err := bkt.SetLifecycle(ctx, bucketName, lifecycle); err != nil {
		log.Fatalf("SetLifecycle failed: %v", err)
	}
	if cfg, err := bkt.GetLifecycle(ctx, bucketName); err == nil {
		log.Printf("Lifecycle config applied. Length=%d bytes", len(cfg))
	} else {
		log.Fatalf("GetLifecycle failed: %v", err)
	}

	// Clean up: remove lifecycle and policy (optional)
	if err := bkt.DeleteLifecycle(ctx, bucketName); err != nil {
		log.Fatalf("DeleteLifecycle failed: %v", err)
	}
	if err := bkt.DeletePolicy(ctx, bucketName); err != nil {
		log.Fatalf("DeletePolicy failed: %v", err)
	}

	log.Printf("Bucket policy and lifecycle demo completed at %s", time.Now().Format(time.RFC3339))
}
