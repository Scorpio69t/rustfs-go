// Package object object/presign_tagging_test.go
package object

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPresignGetAndPut(t *testing.T) {
	service := createAdvancedTestService(t, httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	ctx := context.Background()

	getURL, _, err := service.PresignGet(ctx, "demo-bucket", "demo.txt", 30*time.Second, url.Values{
		"response-content-type": []string{"text/plain"},
	})
	if err != nil {
		t.Fatalf("PresignGet() error = %v", err)
	}
	if !strings.Contains(getURL.RawQuery, "X-Amz-Signature") {
		t.Fatalf("PresignGet() missing signature in URL: %s", getURL.RawQuery)
	}

	putURL, headers, err := service.PresignPut(ctx, "demo-bucket", "upload.txt", 30*time.Second, nil, WithPresignSSES3())
	if err != nil {
		t.Fatalf("PresignPut() error = %v", err)
	}
	if !strings.Contains(putURL.RawQuery, "X-Amz-Signature") {
		t.Fatalf("PresignPut() missing signature in URL: %s", putURL.RawQuery)
	}
	if headers.Get("x-amz-server-side-encryption") != "AES256" {
		t.Fatalf("PresignPut() expected SSE-S3 header, got %q", headers.Get("x-amz-server-side-encryption"))
	}

	headURL, _, err := service.PresignHead(ctx, "demo-bucket", "head.txt", 30*time.Second, url.Values{
		"response-content-disposition": []string{"inline"},
	})
	if err != nil {
		t.Fatalf("PresignHead() error = %v", err)
	}
	if !strings.Contains(headURL.RawQuery, "X-Amz-Signature") {
		t.Fatalf("PresignHead() missing signature in URL: %s", headURL.RawQuery)
	}
}

func TestTaggingCRUD(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.RawQuery, "tagging") {
			t.Errorf("expected tagging query param, got %s", r.URL.RawQuery)
		}
		switch r.Method {
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`<Tagging><TagSet><Tag><Key>env</Key><Value>dev</Value></Tag><Tag><Key>team</Key><Value>storage</Value></Tag></TagSet></Tagging>`)); err != nil {
				t.Fatalf("Failed to write tagging response: %v", err)
			}
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	ctx := context.Background()

	if err := service.SetTagging(ctx, "bucket", "obj", map[string]string{"env": "dev", "team": "storage"}); err != nil {
		t.Fatalf("SetTagging() error = %v", err)
	}

	tags, err := service.GetTagging(ctx, "bucket", "obj")
	if err != nil {
		t.Fatalf("GetTagging() error = %v", err)
	}
	if len(tags) != 2 || tags["env"] != "dev" || tags["team"] != "storage" {
		t.Fatalf("GetTagging() unexpected tags: %+v", tags)
	}

	if err := service.DeleteTagging(ctx, "bucket", "obj"); err != nil {
		t.Fatalf("DeleteTagging() error = %v", err)
	}
}

func TestFPutFGet(t *testing.T) {
	content := []byte("fput/fget content")
	var uploadedBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			if r.Header.Get("x-amz-server-side-encryption") != "AES256" {
				t.Errorf("expected SSE-S3 header on PUT, got %q", r.Header.Get("x-amz-server-side-encryption"))
			}
			body, _ := io.ReadAll(r.Body)
			uploadedBody = body
			w.Header().Set("ETag", `"etag-123"`)
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.Header().Set("Content-Length", "17")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write(content); err != nil {
				t.Fatalf("Failed to write get response: %v", err)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	ctx := context.Background()

	tmpFile := filepath.Join(t.TempDir(), "src.txt")
	if err := os.WriteFile(tmpFile, content, 0o644); err != nil {
		t.Fatalf("failed to write tmp file: %v", err)
	}

	uploadInfo, err := service.FPut(ctx, "bucket", "demo.txt", tmpFile, WithSSES3())
	if err != nil {
		t.Fatalf("FPut() error = %v", err)
	}
	if uploadInfo.ETag == "" {
		t.Fatalf("FPut() missing ETag")
	}
	if len(uploadedBody) == 0 {
		t.Fatalf("FPut() did not send body")
	}

	target := filepath.Join(t.TempDir(), "dst.txt")
	objInfo, err := service.FGet(ctx, "bucket", "demo.txt", target)
	if err != nil {
		t.Fatalf("FGet() error = %v", err)
	}
	if objInfo.Key != "demo.txt" {
		t.Fatalf("FGet() unexpected key %s", objInfo.Key)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(data) != string(content) {
		t.Fatalf("FGet() content mismatch: %s", string(data))
	}
}
