//go:build example

// data_protection.go - bucket versioning, replication, notifications, logging, and version listings
package main

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/Scorpio69t/rustfs-go"
	"github.com/Scorpio69t/rustfs-go/bucket"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/types"
)

func main() {
	const (
		ACCESS_KEY = "rustfsadmin"
		SECRET_KEY = "rustfsadmin"
		ENDPOINT   = "127.0.0.1:9000"
		BUCKET     = "demo-versioned-bucket"
		OBJECT_KEY = "demo.txt"
	)

	client, err := rustfs.New(ENDPOINT, &rustfs.Options{
		Credentials: credentials.NewStaticV4(ACCESS_KEY, SECRET_KEY, ""),
		Secure:      false,
	})
	if err != nil {
		log.Fatalf("init client: %v", err)
	}

	ctx := context.Background()
	bucketSvc := client.Bucket()
	objectSvc := client.Object()

	// Create bucket (idempotent)
	_ = bucketSvc.Create(ctx, BUCKET, bucket.WithRegion("us-east-1"))

	// Enable versioning
	if err := bucketSvc.SetVersioning(ctx, BUCKET, types.VersioningConfig{Status: "Enabled"}); err != nil {
		log.Fatalf("set versioning: %v", err)
	}
	versioning, _ := bucketSvc.GetVersioning(ctx, BUCKET)
	log.Printf("versioning status: %+v", versioning)

	// Configure replication (minimal example payload)
	replicationXML := []byte(`
<ReplicationConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
  <Role>arn:aws:iam::123456789012:role/replication-role</Role>
  <Rule>
    <ID>rule1</ID>
    <Status>Enabled</Status>
    <Prefix></Prefix>
    <Destination>
      <Bucket>arn:aws:s3:::dest-bucket</Bucket>
      <StorageClass>STANDARD</StorageClass>
    </Destination>
  </Rule>
</ReplicationConfiguration>`)
	if err := bucketSvc.SetReplication(ctx, BUCKET, replicationXML); err != nil {
		log.Fatalf("set replication: %v", err)
	}
	log.Println("replication configured")

	// Configure event notifications (optional: requires valid targets)
	// notificationXML := []byte(`
	// <NotificationConfiguration>
	//   <QueueConfiguration>
	//     <Id>queue-events</Id>
	//     <Queue>arn:aws:sqs:::your-valid-queue</Queue>
	//     <Event>s3:ObjectCreated:*</Event>
	//   </QueueConfiguration>
	// </NotificationConfiguration>`)
	// if err := bucketSvc.SetNotification(ctx, BUCKET, notificationXML); err != nil {
	// 	log.Fatalf("set notification: %v", err)
	// }
	// log.Println("notification configured")

	// Configure access logging (optional; skip if backend not implemented)
	// loggingXML := []byte(`
	// <BucketLoggingStatus>
	//   <LoggingEnabled>
	//     <TargetBucket>log-bucket</TargetBucket>
	//     <TargetPrefix>logs/</TargetPrefix>
	//   </LoggingEnabled>
	// </BucketLoggingStatus>`)
	// if err := bucketSvc.SetLogging(ctx, BUCKET, loggingXML); err != nil {
	// 	log.Fatalf("set logging: %v", err)
	// }
	// log.Println("access logging configured")

	// Upload two versions
	_, _ = objectSvc.Put(ctx, BUCKET, OBJECT_KEY, mustReader("first version"), int64(len("first version")))
	time.Sleep(10 * time.Millisecond)
	_, _ = objectSvc.Put(ctx, BUCKET, OBJECT_KEY, mustReader("second version"), int64(len("second version")))

	// List versions (includes delete markers if any)
	log.Println("=== List object versions ===")
	for info := range objectSvc.ListVersions(ctx, BUCKET, object.WithListPrefix(OBJECT_KEY)) {
		if info.Err != nil {
			log.Fatalf("list versions error: %v", info.Err)
		}
		log.Printf("key=%s versionId=%s isLatest=%t deleteMarker=%t size=%d",
			info.Key, info.VersionID, info.IsLatest, info.IsDeleteMarker, info.Size)
	}
}

func mustReader(s string) *strings.Reader {
	return strings.NewReader(s)
}
