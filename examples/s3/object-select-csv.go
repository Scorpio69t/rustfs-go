//go:build example
// +build example

// Example: Select rows from a CSV object
// Demonstrates S3 Select with CSV input/output.
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
		YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
		YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
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

	objectName := "select-sample.csv"
	csvContent := strings.Join([]string{
		"name,age,city",
		"Alice,34,Seattle",
		"Bob,29,Paris",
		"Carol,41,London",
	}, "\n") + "\n"

	// Upload CSV object
	if _, err := objectSvc.Put(ctx, YOURBUCKET, objectName, strings.NewReader(csvContent), int64(len(csvContent)),
		object.WithContentType("text/csv"),
	); err != nil {
		log.Fatalf("Failed to upload CSV object: %v", err)
	}

	// Build select options
	csvIn := s3select.CSVInputOptions{}
	csvIn.SetFileHeaderInfo(s3select.CSVFileHeaderInfoUse)
	csvIn.SetFieldDelimiter(",")

	csvOut := s3select.CSVOutputOptions{}
	csvOut.SetRecordDelimiter("\n")

	opts := s3select.Options{
		Expression:         "SELECT name, age FROM S3Object s WHERE CAST(s.age AS int) > 30",
		ExpressionType:     s3select.QueryExpressionTypeSQL,
		InputSerialization: s3select.InputSerialization{CSV: &csvIn},
		OutputSerialization: s3select.OutputSerialization{
			CSV: &csvOut,
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
