// Package object object/advanced_test.go - 高级功能测试
package object

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

func TestList(t *testing.T) {
	tests := []struct {
		name         string
		bucketName   string
		opts         []ListOption
		responseXML  string
		expectedKeys []string
		wantErr      bool
	}{
		{
			name:       "List objects successfully",
			bucketName: "test-bucket",
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult>
    <Name>test-bucket</Name>
    <Prefix></Prefix>
    <KeyCount>2</KeyCount>
    <MaxKeys>1000</MaxKeys>
    <IsTruncated>false</IsTruncated>
    <Contents>
        <Key>file1.txt</Key>
        <ETag>"abc123"</ETag>
        <Size>1024</Size>
    </Contents>
    <Contents>
        <Key>file2.txt</Key>
        <ETag>"def456"</ETag>
        <Size>2048</Size>
    </Contents>
</ListBucketResult>`,
			expectedKeys: []string{"file1.txt", "file2.txt"},
			wantErr:      false,
		},
		{
			name:       "List with prefix",
			bucketName: "test-bucket",
			opts:       []ListOption{WithListPrefix("docs/")},
			responseXML: `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult>
    <Name>test-bucket</Name>
    <Prefix>docs/</Prefix>
    <KeyCount>1</KeyCount>
    <MaxKeys>1000</MaxKeys>
    <IsTruncated>false</IsTruncated>
    <Contents>
        <Key>docs/readme.md</Key>
        <ETag>"xyz789"</ETag>
        <Size>512</Size>
    </Contents>
</ListBucketResult>`,
			expectedKeys: []string{"docs/readme.md"},
			wantErr:      false,
		},
		{
			name:       "Invalid bucket name",
			bucketName: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(tt.responseXML))
			}))
			defer server.Close()

			service := createAdvancedTestService(t, server)
			ctx := context.Background()
			objectCh := service.List(ctx, tt.bucketName, tt.opts...)

			var keys []string
			var hasError bool
			for obj := range objectCh {
				if obj.Err != nil {
					hasError = true
					break
				}
				keys = append(keys, obj.Key)
			}

			if (hasError) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", hasError, tt.wantErr)
				return
			}

			if !tt.wantErr && len(keys) != len(tt.expectedKeys) {
				t.Errorf("List() returned %d keys, want %d", len(keys), len(tt.expectedKeys))
			}
		})
	}
}

func TestCopy(t *testing.T) {
	tests := []struct {
		name         string
		destBucket   string
		destObject   string
		sourceBucket string
		sourceObject string
		opts         []CopyOption
		statusCode   int
		wantErr      bool
	}{
		{
			name:         "Copy object successfully",
			destBucket:   "dest-bucket",
			destObject:   "dest-object.txt",
			sourceBucket: "source-bucket",
			sourceObject: "source-object.txt",
			statusCode:   http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "Copy with metadata replacement",
			destBucket:   "dest-bucket",
			destObject:   "dest-object.txt",
			sourceBucket: "source-bucket",
			sourceObject: "source-object.txt",
			opts:         []CopyOption{WithCopyMetadata(map[string]string{"author": "test"}, true)},
			statusCode:   http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "Invalid destination bucket",
			destBucket:   "",
			destObject:   "dest-object.txt",
			sourceBucket: "source-bucket",
			sourceObject: "source-object.txt",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				// 验证复制源头
				copySource := r.Header.Get("x-amz-copy-source")
				if copySource == "" && !tt.wantErr {
					t.Error("x-amz-copy-source header not set")
				}
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<CopyObjectResult>
    <ETag>"abc123"</ETag>
    <LastModified>2023-01-01T00:00:00Z</LastModified>
</CopyObjectResult>`))
			}))
			defer server.Close()

			service := createAdvancedTestService(t, server)
			ctx := context.Background()
			_, err := service.Copy(ctx, tt.destBucket, tt.destObject, tt.sourceBucket, tt.sourceObject, tt.opts...)

			if (err != nil) != tt.wantErr {
				t.Errorf("Copy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInitiateMultipartUpload(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		objectName string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "Initiate successfully",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
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
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				if r.URL.Query().Get("uploads") != "" || r.URL.Path == "/test-bucket/test-object.txt" {
					w.Header().Set("Content-Type", "application/xml")
					w.WriteHeader(tt.statusCode)
					w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<InitiateMultipartUploadResult>
    <Bucket>test-bucket</Bucket>
    <Key>test-object.txt</Key>
    <UploadId>test-upload-id-123</UploadId>
</InitiateMultipartUploadResult>`))
				}
			}))
			defer server.Close()

			service := createAdvancedTestService(t, server)
			ctx := context.Background()
			uploadID, err := service.InitiateMultipartUpload(ctx, tt.bucketName, tt.objectName)

			if (err != nil) != tt.wantErr {
				t.Errorf("InitiateMultipartUpload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && uploadID == "" {
				t.Error("InitiateMultipartUpload() returned empty upload ID")
			}
		})
	}
}

func TestUploadPart(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		objectName string
		uploadID   string
		partNumber int
		data       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "Upload part successfully",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			uploadID:   "test-upload-id",
			partNumber: 1,
			data:       "test data",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "Invalid part number",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			uploadID:   "test-upload-id",
			partNumber: 0,
			data:       "test data",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				w.Header().Set("ETag", "\"abc123\"")
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			service := createAdvancedTestService(t, server)
			ctx := context.Background()
			reader := strings.NewReader(tt.data)
			part, err := service.UploadPart(ctx, tt.bucketName, tt.objectName, tt.uploadID, tt.partNumber, reader, int64(len(tt.data)))

			if (err != nil) != tt.wantErr {
				t.Errorf("UploadPart() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && part.ETag == "" {
				t.Error("UploadPart() returned empty ETag")
			}
		})
	}
}

func TestCompleteMultipartUpload(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		objectName string
		uploadID   string
		parts      []types.ObjectPart
		statusCode int
		wantErr    bool
	}{
		{
			name:       "Complete successfully",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			uploadID:   "test-upload-id",
			parts: []types.ObjectPart{
				{PartNumber: 1, ETag: "abc123", Size: 1024},
				{PartNumber: 2, ETag: "def456", Size: 1024},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "Empty parts",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			uploadID:   "test-upload-id",
			parts:      []types.ObjectPart{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/xml")
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<CompleteMultipartUploadResult>
    <Location>http://test-bucket.s3.amazonaws.com/test-object.txt</Location>
    <Bucket>test-bucket</Bucket>
    <Key>test-object.txt</Key>
    <ETag>"abc123-2"</ETag>
</CompleteMultipartUploadResult>`))
			}))
			defer server.Close()

			service := createAdvancedTestService(t, server)
			ctx := context.Background()
			uploadInfo, err := service.CompleteMultipartUpload(ctx, tt.bucketName, tt.objectName, tt.uploadID, tt.parts)

			if (err != nil) != tt.wantErr {
				t.Errorf("CompleteMultipartUpload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && uploadInfo.ETag == "" {
				t.Error("CompleteMultipartUpload() returned empty ETag")
			}
		})
	}
}

func TestAbortMultipartUpload(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		objectName string
		uploadID   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "Abort successfully",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			uploadID:   "test-upload-id",
			statusCode: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "Empty upload ID",
			bucketName: "test-bucket",
			objectName: "test-object.txt",
			uploadID:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			service := createAdvancedTestService(t, server)
			ctx := context.Background()
			err := service.AbortMultipartUpload(ctx, tt.bucketName, tt.objectName, tt.uploadID)

			if (err != nil) != tt.wantErr {
				t.Errorf("AbortMultipartUpload() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// createAdvancedTestService 创建测试用的服务实例
func createAdvancedTestService(t *testing.T, server *httptest.Server) *objectService {
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

	return &objectService{
		executor:      executor,
		locationCache: locationCache,
	}
}
