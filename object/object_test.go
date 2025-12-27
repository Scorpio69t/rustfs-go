// Package object object/object_test.go
package object

import (
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

func TestValidateBucketName(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		wantErr    bool
	}{
		{
			name:       "Valid bucket name",
			bucketName: "test-bucket",
			wantErr:    false,
		},
		{
			name:       "Empty bucket name",
			bucketName: "",
			wantErr:    true,
		},
		{
			name:       "Too short",
			bucketName: "ab",
			wantErr:    true,
		},
		{
			name:       "Too long",
			bucketName: strings.Repeat("a", 64),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBucketName(tt.bucketName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateBucketName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateObjectName(t *testing.T) {
	tests := []struct {
		name       string
		objectName string
		wantErr    bool
	}{
		{
			name:       "Valid object name",
			objectName: "test-object.txt",
			wantErr:    false,
		},
		{
			name:       "Empty object name",
			objectName: "",
			wantErr:    true,
		},
		{
			name:       "Object with path",
			objectName: "folder/subfolder/object.txt",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateObjectName(tt.objectName)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateObjectName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestApplyPutOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []PutOption
		want PutOptions
	}{
		{
			name: "Default options",
			opts: nil,
			want: PutOptions{
				ContentType: "",
			},
		},
		{
			name: "With content type",
			opts: []PutOption{WithContentType("text/plain")},
			want: PutOptions{
				ContentType: "text/plain",
			},
		},
		{
			name: "With metadata",
			opts: []PutOption{
				WithContentType("image/jpeg"),
				WithUserMetadata(map[string]string{"author": "test"}),
			},
			want: PutOptions{
				ContentType:  "image/jpeg",
				UserMetadata: map[string]string{"author": "test"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyPutOptions(tt.opts)
			if got.ContentType != tt.want.ContentType {
				t.Errorf("ContentType = %v, want %v", got.ContentType, tt.want.ContentType)
			}
			if tt.want.UserMetadata != nil {
				if len(got.UserMetadata) != len(tt.want.UserMetadata) {
					t.Errorf("UserMetadata length = %v, want %v", len(got.UserMetadata), len(tt.want.UserMetadata))
				}
			}
		})
	}
}

func TestApplyGetOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []GetOption
		want GetOptions
	}{
		{
			name: "Default options",
			opts: nil,
			want: GetOptions{},
		},
		{
			name: "With range",
			opts: []GetOption{WithGetRange(0, 1023)},
			want: GetOptions{
				RangeStart: 0,
				RangeEnd:   1023,
				SetRange:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyGetOptions(tt.opts)
			if got.SetRange != tt.want.SetRange {
				t.Errorf("SetRange = %v, want %v", got.SetRange, tt.want.SetRange)
			}
			if got.SetRange {
				if got.RangeStart != tt.want.RangeStart || got.RangeEnd != tt.want.RangeEnd {
					t.Errorf("Range = [%d, %d], want [%d, %d]",
						got.RangeStart, got.RangeEnd, tt.want.RangeStart, tt.want.RangeEnd)
				}
			}
		})
	}
}

func TestApplyListOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []ListOption
		want ListOptions
	}{
		{
			name: "Default options",
			opts: nil,
			want: ListOptions{
				MaxKeys: 1000,
			},
		},
		{
			name: "With prefix",
			opts: []ListOption{WithListPrefix("test/")},
			want: ListOptions{
				Prefix:  "test/",
				MaxKeys: 1000,
			},
		},
		{
			name: "With recursive",
			opts: []ListOption{WithListRecursive(true)},
			want: ListOptions{
				Recursive: true,
				MaxKeys:   1000,
			},
		},
		{
			name: "With versions",
			opts: []ListOption{WithListVersions(), WithListMetadata(true), WithListMaxKeys(10)},
			want: ListOptions{
				MaxKeys:      10,
				WithVersions: true,
				WithMetadata: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyListOptions(tt.opts)
			if got.MaxKeys != tt.want.MaxKeys {
				t.Errorf("MaxKeys = %v, want %v", got.MaxKeys, tt.want.MaxKeys)
			}
			if got.Prefix != tt.want.Prefix {
				t.Errorf("Prefix = %v, want %v", got.Prefix, tt.want.Prefix)
			}
			if got.Recursive != tt.want.Recursive {
				t.Errorf("Recursive = %v, want %v", got.Recursive, tt.want.Recursive)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	if service == nil {
		t.Fatal("NewService() returned nil")
	}
}

func BenchmarkApplyPutOptions(b *testing.B) {
	opts := []PutOption{
		WithContentType("text/plain"),
		WithUserMetadata(map[string]string{"key": "value"}),
	}
	for i := 0; i < b.N; i++ {
		_ = applyPutOptions(opts)
	}
}

func BenchmarkApplyListOptions(b *testing.B) {
	opts := []ListOption{
		WithListPrefix("test/"),
		WithListRecursive(true),
		WithListMaxKeys(100),
	}
	for i := 0; i < b.N; i++ {
		_ = applyListOptions(opts)
	}
}
