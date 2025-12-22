// Package bucket bucket/bucket_test.go
package bucket

import (
	"context"
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

func TestCreate(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		opts       []CreateOption
		wantErr    bool
		statusCode int
	}{
		{
			name:       "Create bucket successfully",
			bucketName: "test-bucket",
			opts:       []CreateOption{WithRegion("us-east-1")},
			wantErr:    false,
			statusCode: http.StatusOK,
		},
		{
			name:       "Create bucket with object locking",
			bucketName: "test-bucket-lock",
			opts:       []CreateOption{WithRegion("us-west-2"), WithObjectLocking(true)},
			wantErr:    false,
			statusCode: http.StatusOK,
		},
		{
			name:       "Invalid bucket name",
			bucketName: "",
			opts:       nil,
			wantErr:    true,
		},
		{
			name:       "Bucket name too short",
			bucketName: "ab",
			opts:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.statusCode > 0 {
					w.WriteHeader(tt.statusCode)
				}
			}))
			defer server.Close()

			service := createTestService(t, server)
			err := service.Create(context.Background(), tt.bucketName, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		opts       []DeleteOption
		wantErr    bool
		statusCode int
	}{
		{
			name:       "Delete bucket successfully",
			bucketName: "test-bucket",
			opts:       nil,
			wantErr:    false,
			statusCode: http.StatusNoContent,
		},
		{
			name:       "Delete with force",
			bucketName: "test-bucket",
			opts:       []DeleteOption{WithForceDelete(true)},
			wantErr:    false,
			statusCode: http.StatusOK,
		},
		{
			name:       "Invalid bucket name",
			bucketName: "",
			opts:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.statusCode > 0 {
					w.WriteHeader(tt.statusCode)
				}
			}))
			defer server.Close()

			service := createTestService(t, server)
			err := service.Delete(context.Background(), tt.bucketName, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExists(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		statusCode int
		want       bool
		wantErr    bool
	}{
		{
			name:       "Bucket exists",
			bucketName: "test-bucket",
			statusCode: http.StatusOK,
			want:       true,
			wantErr:    false,
		},
		{
			name:       "Bucket not found",
			bucketName: "nonexistent-bucket",
			statusCode: http.StatusNotFound,
			want:       false,
			wantErr:    false,
		},
		{
			name:       "Invalid bucket name",
			bucketName: "",
			want:       false,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			service := createTestService(t, server)
			got, err := service.Exists(context.Background(), tt.bucketName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Exists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantCount  int
		wantErr    bool
	}{
		{
			name: "List buckets successfully",
			response: `<?xml version="1.0" encoding="UTF-8"?>
<ListAllMyBucketsResult>
    <Owner>
        <ID>owner-id</ID>
        <DisplayName>owner-name</DisplayName>
    </Owner>
    <Buckets>
        <Bucket>
            <Name>bucket1</Name>
            <CreationDate>2023-01-01T00:00:00.000Z</CreationDate>
        </Bucket>
        <Bucket>
            <Name>bucket2</Name>
            <CreationDate>2023-01-02T00:00:00.000Z</CreationDate>
        </Bucket>
    </Buckets>
</ListAllMyBucketsResult>`,
			statusCode: http.StatusOK,
			wantCount:  2,
			wantErr:    false,
		},
		{
			name: "Empty bucket list",
			response: `<?xml version="1.0" encoding="UTF-8"?>
<ListAllMyBucketsResult>
    <Owner>
        <ID>owner-id</ID>
        <DisplayName>owner-name</DisplayName>
    </Owner>
    <Buckets></Buckets>
</ListAllMyBucketsResult>`,
			statusCode: http.StatusOK,
			wantCount:  0,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			service := createTestService(t, server)
			buckets, err := service.List(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(buckets) != tt.wantCount {
				t.Errorf("List() returned %d buckets, want %d", len(buckets), tt.wantCount)
			}
		})
	}
}

func TestGetLocation(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		response   string
		statusCode int
		want       string
		wantErr    bool
	}{
		{
			name:       "Get location successfully",
			bucketName: "test-bucket",
			response:   `<?xml version="1.0" encoding="UTF-8"?><LocationConstraint>us-west-2</LocationConstraint>`,
			statusCode: http.StatusOK,
			want:       "us-west-2",
			wantErr:    false,
		},
		{
			name:       "Empty location (us-east-1)",
			bucketName: "test-bucket",
			response:   `<?xml version="1.0" encoding="UTF-8"?><LocationConstraint></LocationConstraint>`,
			statusCode: http.StatusOK,
			want:       "us-east-1",
			wantErr:    false,
		},
		{
			name:       "Invalid bucket name",
			bucketName: "",
			want:       "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			service := createTestService(t, server)
			got, err := service.GetLocation(context.Background(), tt.bucketName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetLocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("GetLocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
		{
			name:       "Minimum length (3)",
			bucketName: "abc",
			wantErr:    false,
		},
		{
			name:       "Maximum length (63)",
			bucketName: strings.Repeat("a", 63),
			wantErr:    false,
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

func TestApplyCreateOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []CreateOption
		want CreateOptions
	}{
		{
			name: "Default options",
			opts: nil,
			want: CreateOptions{
				Region: "us-east-1",
			},
		},
		{
			name: "With region",
			opts: []CreateOption{WithRegion("us-west-2")},
			want: CreateOptions{
				Region: "us-west-2",
			},
		},
		{
			name: "With object locking",
			opts: []CreateOption{WithObjectLocking(true)},
			want: CreateOptions{
				Region:        "us-east-1",
				ObjectLocking: true,
			},
		},
		{
			name: "Multiple options",
			opts: []CreateOption{
				WithRegion("eu-west-1"),
				WithObjectLocking(true),
				WithForceCreate(true),
			},
			want: CreateOptions{
				Region:        "eu-west-1",
				ObjectLocking: true,
				ForceCreate:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyCreateOptions(tt.opts)
			if got.Region != tt.want.Region ||
				got.ObjectLocking != tt.want.ObjectLocking ||
				got.ForceCreate != tt.want.ForceCreate {
				t.Errorf("applyCreateOptions() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestApplyDeleteOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []DeleteOption
		want DeleteOptions
	}{
		{
			name: "Default options",
			opts: nil,
			want: DeleteOptions{},
		},
		{
			name: "With force delete",
			opts: []DeleteOption{WithForceDelete(true)},
			want: DeleteOptions{
				ForceDelete: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applyDeleteOptions(tt.opts)
			if got.ForceDelete != tt.want.ForceDelete {
				t.Errorf("applyDeleteOptions() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// createTestService Create a test service instance
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

func BenchmarkCreate(b *testing.B) {
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
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.Create(ctx, "test-bucket", WithRegion("us-east-1"))
	}
}

func BenchmarkList(b *testing.B) {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<ListAllMyBucketsResult>
    <Owner><ID>owner-id</ID><DisplayName>owner-name</DisplayName></Owner>
    <Buckets>
        <Bucket><Name>bucket1</Name><CreationDate>2023-01-01T00:00:00.000Z</CreationDate></Bucket>
        <Bucket><Name>bucket2</Name><CreationDate>2023-01-02T00:00:00.000Z</CreationDate></Bucket>
    </Buckets>
</ListAllMyBucketsResult>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		_, _ = service.List(ctx)
	}
}
