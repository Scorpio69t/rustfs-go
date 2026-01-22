//go:build example
// +build example

// Example: Select rows from a JSON object
// Demonstrates S3 Select with JSON input/output.
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
	s3select "github.com/Scorpio69t/rustfs-go/pkg/select"
)

func main() {
	// Connection configuration
	const (
		YOURACCESSKEYID     = "rustfsadmin"
		YOURSECRETACCESSKEY = "rustfsadmin"
		YOURENDPOINT        = "127.0.0.1:9000"
		YOURBUCKET          = "select-bucket"
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
	objectSvc := client.Object()

	// Ensure bucket exists
	exists, err := bucketSvc.Exists(ctx, YOURBUCKET)
	if err != nil {
		log.Fatalf("Failed to check bucket: %v", err)
	}
	if !exists {
		if err := bucketSvc.Create(ctx, YOURBUCKET); err != nil {
			log.Fatalf("Failed to create bucket: %v", err)
		}
		fmt.Printf("Bucket created: %s\n", YOURBUCKET)
	}

	objectName := "select-sample.json"
	jsonLines := strings.Join([]string{
		`{"name":"Alice","age":34,"city":"Seattle"}`,
		`{"name":"Bob","age":29,"city":"Paris"}`,
		`{"name":"Carol","age":41,"city":"London"}`,
	}, "\n") + "\n"

	// Upload JSON object
	if _, err := objectSvc.Put(ctx, YOURBUCKET, objectName, strings.NewReader(jsonLines), int64(len(jsonLines)),
		object.WithContentType("application/json"),
	); err != nil {
		log.Fatalf("Failed to upload JSON object: %v", err)
	}

	// Build select options
	jsonIn := s3select.JSONInputOptions{}
	jsonIn.SetType(s3select.JSONLinesType)

	jsonOut := s3select.JSONOutputOptions{}
	jsonOut.SetRecordDelimiter("\n")

	opts := s3select.Options{
		Expression:         "SELECT s.name, s.age FROM S3Object s WHERE s.age >= 30",
		ExpressionType:     s3select.QueryExpressionTypeSQL,
		InputSerialization: s3select.InputSerialization{JSON: &jsonIn},
		OutputSerialization: s3select.OutputSerialization{
			JSON: &jsonOut,
		},
	}

	results, err := objectSvc.Select(ctx, YOURBUCKET, objectName, opts)
	if err != nil {
		log.Fatalf("Select failed: %v", err)
	}
	defer results.Close()

	data, err := io.ReadAll(results)
	if err != nil {
		log.Fatalf("Failed to read select results: %v", err)
	}

	fmt.Printf("Select results:\n%s\n", string(data))
	if stats := results.Stats(); stats != nil {
		fmt.Printf("Bytes scanned: %d\n", stats.BytesScanned)
		fmt.Printf("Bytes processed: %d\n", stats.BytesProcessed)
		fmt.Printf("Bytes returned: %d\n", stats.BytesReturned)
	}
}
