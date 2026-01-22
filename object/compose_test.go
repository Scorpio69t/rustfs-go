// Package object object/compose_test.go
package object

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestComposeSingleSourceCopy(t *testing.T) {
	const (
		srcBucket = "src-bucket"
		srcObject = "src-object.txt"
		dstBucket = "dst-bucket"
		dstObject = "dst-object.txt"
	)

	var gotCopy bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			w.Header().Set("Content-Length", strconv.FormatInt(1024*1024, 10))
			w.Header().Set("ETag", "\"etag-1\"")
			w.WriteHeader(http.StatusOK)
		case http.MethodPut:
			if r.Header.Get("x-amz-copy-source") == "" {
				t.Error("x-amz-copy-source header not set")
			}
			gotCopy = true
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<CopyObjectResult>
  <ETag>"abc123"</ETag>
  <LastModified>2023-01-01T00:00:00Z</LastModified>
</CopyObjectResult>`)); err != nil {
				t.Fatalf("Failed to write copy response: %v", err)
			}
		default:
			t.Errorf("unexpected method %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	dst := DestinationInfo{Bucket: dstBucket, Object: dstObject}
	sources := []SourceInfo{{Bucket: srcBucket, Object: srcObject}}

	if _, err := service.Compose(context.Background(), dst, sources); err != nil {
		t.Fatalf("Compose() error = %v", err)
	}
	if !gotCopy {
		t.Fatalf("expected compose to use copy for a single small source")
	}
}

func TestComposeMultipleSources(t *testing.T) {
	const (
		srcBucket1 = "src-bucket-1"
		srcObject1 = "src-object-1.txt"
		srcBucket2 = "src-bucket-2"
		srcObject2 = "src-object-2.txt"
		dstBucket  = "dst-bucket"
		dstObject  = "dst-object.txt"
	)

	sizes := map[string]int64{
		"/" + srcBucket1 + "/" + srcObject1: 6 * 1024 * 1024,
		"/" + srcBucket2 + "/" + srcObject2: 6 * 1024 * 1024,
	}
	partCopyCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			size, ok := sizes[r.URL.Path]
			if !ok {
				t.Errorf("unexpected HEAD path %s", r.URL.Path)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Length", strconv.FormatInt(size, 10))
			w.Header().Set("ETag", "\"etag\"")
			w.WriteHeader(http.StatusOK)
		case http.MethodPost:
			if _, ok := r.URL.Query()["uploads"]; ok {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<InitiateMultipartUploadResult>
  <Bucket>` + dstBucket + `</Bucket>
  <Key>` + dstObject + `</Key>
  <UploadId>upload-id-1</UploadId>
</InitiateMultipartUploadResult>`)); err != nil {
					t.Fatalf("Failed to write initiate response: %v", err)
				}
				return
			}
			if r.URL.Query().Get("uploadId") != "" {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<CompleteMultipartUploadResult>
  <Location>http://example.com/` + dstObject + `</Location>
  <Bucket>` + dstBucket + `</Bucket>
  <Key>` + dstObject + `</Key>
  <ETag>"etag-final"</ETag>
</CompleteMultipartUploadResult>`)); err != nil {
					t.Fatalf("Failed to write complete response: %v", err)
				}
				return
			}
			t.Errorf("unexpected POST query %s", r.URL.RawQuery)
			w.WriteHeader(http.StatusBadRequest)
		case http.MethodPut:
			if r.URL.Query().Get("uploadId") == "" || r.URL.Query().Get("partNumber") == "" {
				t.Errorf("unexpected PUT query %s", r.URL.RawQuery)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if r.Header.Get("x-amz-copy-source") == "" {
				t.Error("x-amz-copy-source header not set")
			}
			if r.Header.Get("x-amz-copy-source-range") == "" {
				t.Error("x-amz-copy-source-range header not set")
			}
			partCopyCount++
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<CopyPartResult>
  <ETag>"etag-part"</ETag>
  <LastModified>2023-01-01T00:00:00Z</LastModified>
</CopyPartResult>`)); err != nil {
				t.Fatalf("Failed to write copy part response: %v", err)
			}
		default:
			t.Errorf("unexpected method %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	dst := DestinationInfo{Bucket: dstBucket, Object: dstObject}
	sources := []SourceInfo{
		{Bucket: srcBucket1, Object: srcObject1},
		{Bucket: srcBucket2, Object: srcObject2},
	}

	if _, err := service.Compose(context.Background(), dst, sources); err != nil {
		t.Fatalf("Compose() error = %v", err)
	}
	if partCopyCount != 2 {
		t.Fatalf("expected 2 upload part copy calls, got %d", partCopyCount)
	}
}

func TestComposeConditionalRange(t *testing.T) {
	const (
		srcBucket = "src-bucket"
		srcObject = "src-object.txt"
		dstBucket = "dst-bucket"
		dstObject = "dst-object.txt"
	)

	var sawConditional bool
	var sawRange bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			w.Header().Set("Content-Length", strconv.FormatInt(100, 10))
			w.Header().Set("ETag", "\"etag-conditional\"")
			w.WriteHeader(http.StatusOK)
		case http.MethodPost:
			if _, ok := r.URL.Query()["uploads"]; ok {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<InitiateMultipartUploadResult>
  <Bucket>` + dstBucket + `</Bucket>
  <Key>` + dstObject + `</Key>
  <UploadId>upload-id-2</UploadId>
</InitiateMultipartUploadResult>`)); err != nil {
					t.Fatalf("Failed to write initiate response: %v", err)
				}
				return
			}
			if r.URL.Query().Get("uploadId") != "" {
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<CompleteMultipartUploadResult>
  <Location>http://example.com/` + dstObject + `</Location>
  <Bucket>` + dstBucket + `</Bucket>
  <Key>` + dstObject + `</Key>
  <ETag>"etag-final"</ETag>
</CompleteMultipartUploadResult>`)); err != nil {
					t.Fatalf("Failed to write complete response: %v", err)
				}
				return
			}
			t.Errorf("unexpected POST query %s", r.URL.RawQuery)
			w.WriteHeader(http.StatusBadRequest)
		case http.MethodPut:
			if r.URL.Query().Get("uploadId") == "" || r.URL.Query().Get("partNumber") == "" {
				t.Errorf("unexpected PUT query %s", r.URL.RawQuery)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if r.Header.Get("x-amz-copy-source-if-match") == "\"etag-conditional\"" {
				sawConditional = true
			}
			if r.Header.Get("x-amz-copy-source-range") == "bytes=1-10" {
				sawRange = true
			}
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<CopyPartResult>
  <ETag>"etag-part"</ETag>
  <LastModified>2023-01-01T00:00:00Z</LastModified>
</CopyPartResult>`)); err != nil {
				t.Fatalf("Failed to write copy part response: %v", err)
			}
		default:
			t.Errorf("unexpected method %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	dst := DestinationInfo{Bucket: dstBucket, Object: dstObject}
	sources := []SourceInfo{{
		Bucket:     srcBucket,
		Object:     srcObject,
		RangeStart: 1,
		RangeEnd:   10,
		RangeSet:   true,
		MatchETag:  "\"etag-conditional\"",
	}}

	if _, err := service.Compose(context.Background(), dst, sources); err != nil {
		t.Fatalf("Compose() error = %v", err)
	}
	if !sawConditional {
		t.Fatalf("expected conditional header for compose copy")
	}
	if !sawRange {
		t.Fatalf("expected range header for compose copy")
	}
}
