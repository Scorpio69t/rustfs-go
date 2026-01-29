// Package object object/multipart_list_test.go
package object

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListMultipartUploads(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ListMultipartUploadsResult>
  <Bucket>demo-bucket</Bucket>
  <KeyMarker></KeyMarker>
  <UploadIdMarker></UploadIdMarker>
  <NextKeyMarker></NextKeyMarker>
  <NextUploadIdMarker></NextUploadIdMarker>
  <MaxUploads>2</MaxUploads>
  <IsTruncated>false</IsTruncated>
  <Upload>
    <Key>obj1.txt</Key>
    <UploadId>upload-1</UploadId>
    <Initiator>
      <ID>init-id</ID>
      <DisplayName>init</DisplayName>
    </Initiator>
    <Owner>
      <ID>owner-id</ID>
      <DisplayName>owner</DisplayName>
    </Owner>
    <StorageClass>STANDARD</StorageClass>
    <Initiated>2024-01-01T00:00:00.000Z</Initiated>
  </Upload>
</ListMultipartUploadsResult>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if _, ok := r.URL.Query()["uploads"]; !ok {
			t.Fatalf("expected uploads query param")
		}
		if got := r.URL.Query().Get("max-uploads"); got != "2" {
			t.Fatalf("expected max-uploads=2, got %q", got)
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(responseXML)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	result, err := service.ListMultipartUploads(context.Background(), "demo-bucket", WithMultipartMaxUploads(2))
	if err != nil {
		t.Fatalf("ListMultipartUploads() error = %v", err)
	}
	if len(result.Uploads) != 1 {
		t.Fatalf("expected 1 upload, got %d", len(result.Uploads))
	}
	if result.Uploads[0].UploadID != "upload-1" {
		t.Fatalf("unexpected upload ID %q", result.Uploads[0].UploadID)
	}
}

func TestListObjectParts(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<ListPartsResult>
  <Bucket>demo-bucket</Bucket>
  <Key>obj1.txt</Key>
  <UploadId>upload-1</UploadId>
  <PartNumberMarker>0</PartNumberMarker>
  <NextPartNumberMarker>1</NextPartNumberMarker>
  <MaxParts>2</MaxParts>
  <IsTruncated>false</IsTruncated>
  <Part>
    <PartNumber>1</PartNumber>
    <ETag>"etag-1"</ETag>
    <Size>5</Size>
    <LastModified>2024-01-01T00:00:00.000Z</LastModified>
  </Part>
</ListPartsResult>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if got := r.URL.Query().Get("uploadId"); got != "upload-1" {
			t.Fatalf("expected uploadId=upload-1, got %q", got)
		}
		if got := r.URL.Query().Get("max-parts"); got != "2" {
			t.Fatalf("expected max-parts=2, got %q", got)
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(responseXML)); err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	result, err := service.ListObjectParts(context.Background(), "demo-bucket", "obj1.txt", "upload-1", WithListPartsMax(2))
	if err != nil {
		t.Fatalf("ListObjectParts() error = %v", err)
	}
	if len(result.Parts) != 1 {
		t.Fatalf("expected 1 part, got %d", len(result.Parts))
	}
	if result.Parts[0].PartNumber != 1 {
		t.Fatalf("unexpected part number %d", result.Parts[0].PartNumber)
	}
}
