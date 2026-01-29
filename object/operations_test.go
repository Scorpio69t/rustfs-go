// Package object object/operations_test.go
package object

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/Scorpio69t/rustfs-go/internal/cache"
	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/pkg/cse"
	"github.com/Scorpio69t/rustfs-go/types"
)

func TestPut(t *testing.T) {
	cseClient, err := cse.New(bytes.Repeat([]byte{0x11}, 32))
	if err != nil {
		t.Fatalf("Failed to create cse client: %v", err)
	}

	tests := []struct {
		name       string
		bucketName string
		objectName string
		data       string
		opts       []PutOption
		wantChecksumMode      string
		wantChecksumAlgorithm string
		wantCSE bool
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
			name:       "Put with checksum mode",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			data:       "checksum data",
			opts: []PutOption{
				WithChecksumMode("ENABLED"),
				WithChecksumAlgorithm("CRC32C"),
			},
			wantChecksumMode:      "ENABLED",
			wantChecksumAlgorithm: "CRC32C",
			statusCode:            http.StatusOK,
			wantErr:               false,
		},
		{
			name:       "Put with client-side encryption",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			data:       "secret data",
			opts:       []PutOption{WithPutCSE(cseClient)},
			wantCSE:    true,
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
				if tt.wantChecksumMode != "" {
					if got := r.Header.Get("x-amz-checksum-mode"); got != tt.wantChecksumMode {
						t.Errorf("checksum mode = %q, want %q", got, tt.wantChecksumMode)
					}
				}
				if tt.wantChecksumAlgorithm != "" {
					if got := r.Header.Get("x-amz-checksum-algorithm"); got != tt.wantChecksumAlgorithm {
						t.Errorf("checksum algorithm = %q, want %q", got, tt.wantChecksumAlgorithm)
					}
				}
				if tt.wantCSE {
					if got := r.Header.Get("x-amz-meta-rustfs-cse-algorithm"); got == "" {
						t.Error("missing cse algorithm metadata header")
					}
					if got := r.Header.Get("x-amz-meta-rustfs-cse-nonce"); got == "" {
						t.Error("missing cse nonce metadata header")
					}
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Fatalf("Failed to read request body: %v", err)
					}
					if len(body) <= len(tt.data) {
						t.Errorf("encrypted payload size = %d, want > %d", len(body), len(tt.data))
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
	cseClient, err := cse.New(bytes.Repeat([]byte{0x22}, 32))
	if err != nil {
		t.Fatalf("Failed to create cse client: %v", err)
	}

	tests := []struct {
		name       string
		bucketName string
		objectName string
		opts       []GetOption
		wantQuery  url.Values
		wantRange  bool
		response   string
		cseClient  *cse.Client
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
			wantRange:  true,
			response:   "Hello",
			statusCode: http.StatusPartialContent,
			wantErr:    false,
		},
		{
			name:       "Get with response overrides",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			opts: []GetOption{WithGetResponseHeaders(url.Values{
				"response-content-type":        []string{"text/plain"},
				"response-content-disposition": []string{"inline"},
			})},
			wantQuery: url.Values{
				"response-content-type":        []string{"text/plain"},
				"response-content-disposition": []string{"inline"},
			},
			response:   "Hello",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "Get with client-side encryption",
			bucketName: "test-bucket",
			objectName: "secret.txt",
			opts:       []GetOption{WithGetCSE(cseClient)},
			response:   "Encrypted payload",
			cseClient:  cseClient,
			statusCode: http.StatusOK,
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
				if tt.wantRange {
					rangeHeader := r.Header.Get("Range")
					if rangeHeader == "" {
						t.Error("Range header not set")
					}
				}
				if len(tt.wantQuery) > 0 {
					for key, values := range tt.wantQuery {
						got := r.URL.Query()[key]
						if len(got) != len(values) {
							t.Errorf("Expected query %s=%v, got %v", key, values, got)
							continue
						}
						for i, v := range values {
							if got[i] != v {
								t.Errorf("Expected query %s=%v, got %v", key, values, got)
								break
							}
						}
					}
				}
				body := []byte(tt.response)
				if tt.cseClient != nil {
					encrypted, metadata, err := tt.cseClient.Encrypt(strings.NewReader(tt.response))
					if err != nil {
						t.Fatalf("Failed to encrypt payload: %v", err)
					}
					body = encrypted
					w.Header().Set("x-amz-meta-rustfs-cse-algorithm", metadata["rustfs-cse-algorithm"])
					w.Header().Set("x-amz-meta-rustfs-cse-nonce", metadata["rustfs-cse-nonce"])
				}
				w.Header().Set("Content-Type", "text/plain")
				w.Header().Set("ETag", "\"abc123\"")
				w.Header().Set("Content-Length", strconv.Itoa(len(body)))
				w.WriteHeader(tt.statusCode)
				if _, err := w.Write(body); err != nil {
					t.Fatalf("Failed to write response: %v", err)
				}
			}))
			defer server.Close()

			service := createTestService(t, server)
			reader, _, err := service.Get(context.Background(), tt.bucketName, tt.objectName, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && reader != nil {
				defer func() {
					if err := reader.Close(); err != nil {
						t.Fatalf("Failed to close reader: %v", err)
					}
				}()
				data, err := io.ReadAll(reader)
				if err != nil {
					t.Fatalf("Failed to read response: %v", err)
				}
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
		if _, err := w.Write([]byte(response)); err != nil {
			b.Fatalf("Failed to write response: %v", err)
		}
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
			if _, err := io.Copy(io.Discard, reader); err != nil {
				b.Fatalf("Failed to discard response: %v", err)
			}
			if err := reader.Close(); err != nil {
				b.Fatalf("Failed to close reader: %v", err)
			}
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
