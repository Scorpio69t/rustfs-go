// Package object object/append_test.go
package object

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
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

func TestAppendLargeObject(t *testing.T) {
	const offset = int64(1024)
	const size = int64(5 * 1024 * 1024)
	expectedSize := offset + size
	payload := bytes.Repeat([]byte("a"), int(size))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if got := r.Header.Get("x-amz-write-offset-bytes"); got != strconv.FormatInt(offset, 10) {
			t.Errorf("unexpected offset header %q", got)
		}
		if r.ContentLength != size {
			t.Errorf("unexpected Content-Length %d", r.ContentLength)
		}
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("ETag", "\"etag-large\"")
		w.Header().Set("x-amz-object-size", strconv.FormatInt(expectedSize, 10))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	uploadInfo, err := service.Append(context.Background(), "bucket", "object", bytes.NewReader(payload), size, offset)
	if err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	if uploadInfo.Size != expectedSize {
		t.Fatalf("Append() size = %d, want %d", uploadInfo.Size, expectedSize)
	}
}

func TestAppendConcurrent(t *testing.T) {
	const workers = 4

	var mu sync.Mutex
	seenOffsets := make(map[int64]struct{})
	var handlerErrs []string

	recordErr := func(msg string) {
		mu.Lock()
		handlerErrs = append(handlerErrs, msg)
		mu.Unlock()
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			recordErr(fmt.Sprintf("expected PUT request, got %s", r.Method))
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		offsetStr := r.Header.Get("x-amz-write-offset-bytes")
		if offsetStr == "" {
			recordErr("missing x-amz-write-offset-bytes header")
		}
		offset, err := strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			recordErr(fmt.Sprintf("invalid offset header: %v", err))
		}
		mu.Lock()
		seenOffsets[offset] = struct{}{}
		mu.Unlock()

		_, _ = io.Copy(io.Discard, r.Body)
		finalSize := offset + r.ContentLength
		w.Header().Set("ETag", "\"etag-concurrent\"")
		w.Header().Set("x-amz-object-size", strconv.FormatInt(finalSize, 10))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	ctx := context.Background()

	errCh := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			size := int64(1024 + i)
			offset := int64(i * 2048)
			payload := bytes.Repeat([]byte("b"), int(size))
			uploadInfo, err := service.Append(ctx, "bucket", "object", bytes.NewReader(payload), size, offset)
			if err != nil {
				errCh <- err
				return
			}
			if uploadInfo.Size != offset+size {
				errCh <- fmt.Errorf("append size %d, want %d", uploadInfo.Size, offset+size)
				return
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("Append() error = %v", err)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(handlerErrs) > 0 {
		t.Fatalf("handler errors: %v", handlerErrs)
	}
	if len(seenOffsets) != workers {
		t.Fatalf("expected %d offsets, got %d", workers, len(seenOffsets))
	}
}
