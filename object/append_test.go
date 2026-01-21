// Package object object/append_test.go
package object

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestAppendWithExplicitOffset(t *testing.T) {
	data := []byte("hello")
	offset := int64(0)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if got := r.Header.Get("x-amz-write-offset-bytes"); got != strconv.FormatInt(offset, 10) {
			t.Errorf("unexpected offset header %q", got)
		}
		w.Header().Set("ETag", "\"etag-append\"")
		w.Header().Set("x-amz-object-size", strconv.FormatInt(int64(len(data)), 10))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	uploadInfo, err := service.Append(context.Background(), "bucket", "object", bytes.NewReader(data), int64(len(data)), offset)
	if err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if uploadInfo.Size != int64(len(data)) {
		t.Fatalf("Append() size = %d, want %d", uploadInfo.Size, len(data))
	}
}

func TestAppendWithAutoOffset(t *testing.T) {
	data := []byte("world")
	initialSize := int64(7)
	expectedSize := initialSize + int64(len(data))

	var sawHead bool
	var sawPut bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			sawHead = true
			w.Header().Set("Content-Length", strconv.FormatInt(initialSize, 10))
			w.WriteHeader(http.StatusOK)
		case http.MethodPut:
			sawPut = true
			if got := r.Header.Get("x-amz-write-offset-bytes"); got != strconv.FormatInt(initialSize, 10) {
				t.Errorf("unexpected offset header %q", got)
			}
			w.Header().Set("ETag", "\"etag-append\"")
			w.Header().Set("x-amz-object-size", strconv.FormatInt(expectedSize, 10))
			w.WriteHeader(http.StatusOK)
		default:
			t.Errorf("unexpected method %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	uploadInfo, err := service.Append(context.Background(), "bucket", "object", bytes.NewReader(data), int64(len(data)), -1)
	if err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if !sawHead || !sawPut {
		t.Fatalf("expected both HEAD and PUT requests")
	}
	if uploadInfo.Size != expectedSize {
		t.Fatalf("Append() size = %d, want %d", uploadInfo.Size, expectedSize)
	}
}

func TestAppendMissingSizeHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", "\"etag-append\"")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	_, err := service.Append(context.Background(), "bucket", "object", bytes.NewReader([]byte("data")), 4, 0)
	if err == nil {
		t.Fatalf("expected Append() to fail when size header is missing")
	}
}
