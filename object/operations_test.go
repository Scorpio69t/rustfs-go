// Package object object/operations_test.go
package object

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Scorpio69t/rustfs-go/internal/cache"
	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/types"
)

func TestPut(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		objectName string
		data       string
		opts       []PutOption
		statusCode int
		wantErr    bool
	}{
		{
			name:       "Put object successfully",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			data:       "Hello, World!",
			opts:       []PutOption{WithContentType("text/plain")},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "Put with metadata",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			data:       "test data",
			opts:       []PutOption{WithUserMetadata(map[string]string{"author": "test"})},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "Invalid bucket name",
			bucketName: "",
			objectName: "test-object.txt",
			data:       "test data",
			wantErr:    true,
		},
		{
			name:       "Invalid object name",
			bucketName: "test-bucket",
			objectName: "",
			data:       "test data",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// verify request method
				if r.Method != http.MethodPut {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				// verify Content-Type
				if len(tt.opts) > 0 {
					contentType := r.Header.Get("Content-Type")
					if contentType == "" {
						t.Error("Content-Type header not set")
					}
				}
				w.Header().Set("ETag", "\"abc123\"")
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			service := createTestService(t, server)
			reader := strings.NewReader(tt.data)
			_, err := service.Put(context.Background(), tt.bucketName, tt.objectName, reader, int64(len(tt.data)), tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Put() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		objectName string
		opts       []GetOption
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "Get object successfully",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			response:   "Hello, World!",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "Get with range",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			opts:       []GetOption{WithGetRange(0, 10)},
			response:   "Hello",
			statusCode: http.StatusPartialContent,
			wantErr:    false,
		},
		{
			name:       "Invalid bucket name",
			bucketName: "",
			objectName: "test-object.txt",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// verify request method
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				// verify Range header
				if len(tt.opts) > 0 {
					rangeHeader := r.Header.Get("Range")
					if rangeHeader == "" {
						t.Error("Range header not set")
					}
				}
				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("ETag", "\"abc123\"")
				w.Header().Set("Content-Length", string(rune(len(tt.response))))
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			service := createTestService(t, server)
			reader, _, err := service.Get(context.Background(), tt.bucketName, tt.objectName, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && reader != nil {
				defer reader.Close()
				data, _ := io.ReadAll(reader)
				if string(data) != tt.response {
					t.Errorf("Get() data = %s, want %s", string(data), tt.response)
				}
			}
		})
	}
}

func TestStat(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		objectName string
		opts       []StatOption
		statusCode int
		wantErr    bool
	}{
		{
			name:       "Stat object successfully",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "Object not found",
			bucketName: "test-bucket",
			objectName: "nonexistent.txt",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "Invalid bucket name",
			bucketName: "",
			objectName: "test-object.txt",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// verify request method
				if r.Method != http.MethodHead {
					t.Errorf("Expected HEAD request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("Content-Length", "1024")
				w.Header().Set("ETag", "\"abc123\"")
				w.Header().Set("Last-Modified", "Mon, 02 Jan 2006 15:04:05 GMT")
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			service := createTestService(t, server)
			info, err := service.Stat(context.Background(), tt.bucketName, tt.objectName, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Stat() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if info.Key != tt.objectName {
					t.Errorf("Stat() Key = %s, want %s", info.Key, tt.objectName)
				}
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		objectName string
		opts       []DeleteOption
		statusCode int
		wantErr    bool
	}{
		{
			name:       "Delete object successfully",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			statusCode: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "Delete with 200 OK",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "Object not found (still success)",
			bucketName: "test-bucket",
			objectName: "nonexistent.txt",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "Invalid bucket name",
			bucketName: "",
			objectName: "test-object.txt",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// verify request method
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			service := createTestService(t, server)
			err := service.Delete(context.Background(), tt.bucketName, tt.objectName, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkPut(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", "\"abc123\"")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	creds := credentials.NewStaticV4("access-key", "secret-key", "")
	locationCache := cache.NewLocationCache(0)

	executor := core.NewExecutor(core.ExecutorConfig{
		HTTPClient:   server.Client(),
		EndpointURL:  serverURL,
		Credentials:  creds,
		Region:       "us-east-1",
		BucketLookup: int(types.BucketLookupPath),
		MaxRetries:   1,
	})

	service := NewService(executor, locationCache)
	ctx := context.Background()
	data := []byte("test data")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := bytes.NewReader(data)
		_, _ = service.Put(ctx, "test-bucket", "test-object.txt", reader, int64(len(data)))
	}
}

func BenchmarkGet(b *testing.B) {
	response := "Hello, World!"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("ETag", "\"abc123\"")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(response))
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	creds := credentials.NewStaticV4("access-key", "secret-key", "")
	locationCache := cache.NewLocationCache(0)

	executor := core.NewExecutor(core.ExecutorConfig{
		HTTPClient:   server.Client(),
		EndpointURL:  serverURL,
		Credentials:  creds,
		Region:       "us-east-1",
		BucketLookup: int(types.BucketLookupPath),
		MaxRetries:   1,
	})

	service := NewService(executor, locationCache)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader, _, _ := service.Get(ctx, "test-bucket", "test-object.txt")
		if reader != nil {
			io.Copy(io.Discard, reader)
			reader.Close()
		}
	}
}

// createTestService creates a service instance for testing
func createTestService(t *testing.T, server *httptest.Server) Service {
	t.Helper()

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Failed to parse server URL: %v", err)
	}

	creds := credentials.NewStaticV4("access-key", "secret-key", "")
	locationCache := cache.NewLocationCache(0)

	executor := core.NewExecutor(core.ExecutorConfig{
		HTTPClient:   server.Client(),
		EndpointURL:  serverURL,
		Credentials:  creds,
		Region:       "us-east-1",
		BucketLookup: int(types.BucketLookupPath),
		MaxRetries:   1,
	})

	return NewService(executor, locationCache)
}
