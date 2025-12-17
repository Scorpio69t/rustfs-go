// Package core internal/core/executor_test.go
package core

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/types"
)

func TestNewExecutor(t *testing.T) {
	endpointURL, _ := url.Parse("https://s3.amazonaws.com")
	creds := credentials.NewStaticV4("access-key", "secret-key", "")

	config := ExecutorConfig{
		HTTPClient:  &http.Client{},
		EndpointURL: endpointURL,
		Credentials: creds,
		Region:      "us-east-1",
		Secure:      true,
		MaxRetries:  5,
	}

	executor := NewExecutor(config)

	if executor == nil {
		t.Fatal("NewExecutor() returned nil")
	}

	if executor.maxRetries != 5 {
		t.Errorf("maxRetries = %d, want 5", executor.maxRetries)
	}

	if executor.region != "us-east-1" {
		t.Errorf("region = %s, want us-east-1", executor.region)
	}
}

func TestNewExecutorDefaultRetries(t *testing.T) {
	endpointURL, _ := url.Parse("https://s3.amazonaws.com")
	creds := credentials.NewStaticV4("access-key", "secret-key", "")

	config := ExecutorConfig{
		HTTPClient:  &http.Client{},
		EndpointURL: endpointURL,
		Credentials: creds,
	}

	executor := NewExecutor(config)

	if executor.maxRetries != 10 {
		t.Errorf("maxRetries = %d, want 10 (default)", executor.maxRetries)
	}
}

func TestMakeTargetURL(t *testing.T) {
	tests := []struct {
		name         string
		endpointURL  string
		bucketLookup int
		bucketName   string
		objectName   string
		queryValues  url.Values
		want         string
	}{
		{
			name:         "Path style - bucket only",
			endpointURL:  "https://s3.amazonaws.com",
			bucketLookup: int(types.BucketLookupPath),
			bucketName:   "my-bucket",
			objectName:   "",
			want:         "https://s3.amazonaws.com/my-bucket/",
		},
		{
			name:         "Path style - bucket and object",
			endpointURL:  "https://s3.amazonaws.com",
			bucketLookup: int(types.BucketLookupPath),
			bucketName:   "my-bucket",
			objectName:   "my-object.txt",
			want:         "https://s3.amazonaws.com/my-bucket/my-object.txt",
		},
		{
			name:         "Virtual host style - bucket only",
			endpointURL:  "https://s3.amazonaws.com",
			bucketLookup: int(types.BucketLookupDNS),
			bucketName:   "my-bucket",
			objectName:   "",
			want:         "https://my-bucket.s3.amazonaws.com/",
		},
		{
			name:         "Virtual host style - bucket and object",
			endpointURL:  "https://s3.amazonaws.com",
			bucketLookup: int(types.BucketLookupDNS),
			bucketName:   "my-bucket",
			objectName:   "my-object.txt",
			want:         "https://my-bucket.s3.amazonaws.com/my-object.txt",
		},
		{
			name:         "Object with special characters",
			endpointURL:  "https://s3.amazonaws.com",
			bucketLookup: int(types.BucketLookupPath),
			bucketName:   "my-bucket",
			objectName:   "folder/my object+test.txt",
			want:         "https://s3.amazonaws.com/my-bucket/folder/my%20object%2Btest.txt",
		},
		{
			name:         "With query parameters",
			endpointURL:  "https://s3.amazonaws.com",
			bucketLookup: int(types.BucketLookupPath),
			bucketName:   "my-bucket",
			objectName:   "my-object.txt",
			queryValues:  url.Values{"max-keys": []string{"100"}, "prefix": []string{"test/"}},
			want:         "https://s3.amazonaws.com/my-bucket/my-object.txt?max-keys=100&prefix=test%2F",
		},
		{
			name:         "No bucket",
			endpointURL:  "https://s3.amazonaws.com",
			bucketLookup: int(types.BucketLookupPath),
			bucketName:   "",
			objectName:   "",
			want:         "https://s3.amazonaws.com/",
		},
		{
			name:         "HTTP with port 80 (should remove)",
			endpointURL:  "http://localhost:80",
			bucketLookup: int(types.BucketLookupPath),
			bucketName:   "my-bucket",
			objectName:   "",
			want:         "http://localhost/my-bucket/",
		},
		{
			name:         "HTTPS with port 443 (should remove)",
			endpointURL:  "https://s3.amazonaws.com:443",
			bucketLookup: int(types.BucketLookupPath),
			bucketName:   "my-bucket",
			objectName:   "",
			want:         "https://s3.amazonaws.com/my-bucket/",
		},
		{
			name:         "Custom port (should keep)",
			endpointURL:  "http://localhost:9000",
			bucketLookup: int(types.BucketLookupPath),
			bucketName:   "my-bucket",
			objectName:   "",
			want:         "http://localhost:9000/my-bucket/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			endpointURL, err := url.Parse(tt.endpointURL)
			if err != nil {
				t.Fatalf("url.Parse() error = %v", err)
			}

			executor := &Executor{
				endpointURL:  endpointURL,
				bucketLookup: tt.bucketLookup,
			}

			got, err := executor.makeTargetURL(tt.bucketName, tt.objectName, "", tt.queryValues)
			if err != nil {
				t.Fatalf("makeTargetURL() error = %v", err)
			}

			if got.String() != tt.want {
				t.Errorf("makeTargetURL() = %s, want %s", got.String(), tt.want)
			}
		})
	}
}

func TestIsVirtualHostStyleRequest(t *testing.T) {
	tests := []struct {
		name         string
		bucketLookup int
		bucketName   string
		https        bool
		want         bool
	}{
		{
			name:         "DNS lookup",
			bucketLookup: int(types.BucketLookupDNS),
			bucketName:   "my-bucket",
			want:         true,
		},
		{
			name:         "Path lookup",
			bucketLookup: int(types.BucketLookupPath),
			bucketName:   "my-bucket",
			want:         false,
		},
		{
			name:         "Auto - valid bucket",
			bucketLookup: int(types.BucketLookupAuto),
			bucketName:   "my-bucket",
			https:        false,
			want:         true,
		},
		{
			name:         "Auto - bucket with dot in HTTPS",
			bucketLookup: int(types.BucketLookupAuto),
			bucketName:   "my.bucket",
			https:        true,
			want:         false,
		},
		{
			name:         "Auto - bucket with dot in HTTP",
			bucketLookup: int(types.BucketLookupAuto),
			bucketName:   "my.bucket",
			https:        false,
			want:         true,
		},
		{
			name:         "Empty bucket",
			bucketLookup: int(types.BucketLookupDNS),
			bucketName:   "",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scheme := "http"
			if tt.https {
				scheme = "https"
			}
			endpointURL, _ := url.Parse(scheme + "://s3.amazonaws.com")

			executor := &Executor{
				endpointURL:  endpointURL,
				bucketLookup: tt.bucketLookup,
			}

			got := executor.isVirtualHostStyleRequest(tt.bucketName)
			if got != tt.want {
				t.Errorf("isVirtualHostStyleRequest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidVirtualHostBucket(t *testing.T) {
	tests := []struct {
		name       string
		bucketName string
		https      bool
		want       bool
	}{
		{
			name:       "Valid bucket name",
			bucketName: "my-bucket",
			https:      false,
			want:       true,
		},
		{
			name:       "Bucket with dot - HTTP",
			bucketName: "my.bucket",
			https:      false,
			want:       true,
		},
		{
			name:       "Bucket with dot - HTTPS",
			bucketName: "my.bucket",
			https:      true,
			want:       false,
		},
		{
			name:       "Too short bucket name",
			bucketName: "ab",
			https:      false,
			want:       false,
		},
		{
			name:       "Too long bucket name",
			bucketName: strings.Repeat("a", 64),
			https:      false,
			want:       false,
		},
		{
			name:       "IP address bucket name",
			bucketName: "192.168.1.1",
			https:      false,
			want:       false,
		},
		{
			name:       "Valid 63 char bucket",
			bucketName: strings.Repeat("a", 63),
			https:      false,
			want:       true,
		},
		{
			name:       "Valid 3 char bucket",
			bucketName: "abc",
			https:      false,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidVirtualHostBucket(tt.bucketName, tt.https)
			if got != tt.want {
				t.Errorf("isValidVirtualHostBucket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldRetry(t *testing.T) {
	executor := &Executor{
		maxRetries: 3,
	}

	tests := []struct {
		name    string
		err     error
		attempt int
		want    bool
	}{
		{
			name:    "Nil error",
			err:     nil,
			attempt: 0,
			want:    false,
		},
		{
			name:    "Max attempts reached",
			err:     &url.Error{Op: "Get", Err: io.EOF},
			attempt: 2,
			want:    false,
		},
		{
			name:    "Connection refused",
			err:     &url.Error{Op: "Get", Err: &testNetError{msg: "connection refused", temp: true}},
			attempt: 0,
			want:    true,
		},
		{
			name:    "Timeout error",
			err:     &testNetError{msg: "i/o timeout", timeout: true},
			attempt: 0,
			want:    true,
		},
		{
			name:    "Temporary error",
			err:     &testNetError{msg: "temporary error", temp: true},
			attempt: 0,
			want:    true,
		},
		{
			name:    "Non-retryable error",
			err:     io.EOF,
			attempt: 0,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := executor.shouldRetry(tt.err, tt.attempt)
			if got != tt.want {
				t.Errorf("shouldRetry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldRetryResponse(t *testing.T) {
	executor := &Executor{
		maxRetries: 3,
	}

	tests := []struct {
		name       string
		statusCode int
		attempt    int
		want       bool
	}{
		{
			name:       "500 Internal Server Error",
			statusCode: 500,
			attempt:    0,
			want:       true,
		},
		{
			name:       "502 Bad Gateway",
			statusCode: 502,
			attempt:    0,
			want:       true,
		},
		{
			name:       "503 Service Unavailable",
			statusCode: 503,
			attempt:    0,
			want:       true,
		},
		{
			name:       "429 Too Many Requests",
			statusCode: 429,
			attempt:    0,
			want:       true,
		},
		{
			name:       "200 OK",
			statusCode: 200,
			attempt:    0,
			want:       false,
		},
		{
			name:       "404 Not Found",
			statusCode: 404,
			attempt:    0,
			want:       false,
		},
		{
			name:       "Max attempts reached",
			statusCode: 500,
			attempt:    2,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
			}
			got := executor.shouldRetryResponse(resp, tt.attempt)
			if got != tt.want {
				t.Errorf("shouldRetryResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExecuteSuccess(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	creds := credentials.NewStaticV4("access-key", "secret-key", "")

	executor := NewExecutor(ExecutorConfig{
		HTTPClient:   server.Client(),
		EndpointURL:  serverURL,
		Credentials:  creds,
		MaxRetries:   3,
		BucketLookup: int(types.BucketLookupPath), // 使用路径风格避免 DNS 查找
	})

	req := NewRequest(context.Background(), http.MethodGet, RequestMetadata{
		BucketName: "test-bucket",
	})

	resp, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestExecuteRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	creds := credentials.NewStaticV4("access-key", "secret-key", "")

	executor := NewExecutor(ExecutorConfig{
		HTTPClient:   server.Client(),
		EndpointURL:  serverURL,
		Credentials:  creds,
		MaxRetries:   5,
		BucketLookup: int(types.BucketLookupPath), // 使用路径风格避免 DNS 查找
	})

	req := NewRequest(context.Background(), http.MethodGet, RequestMetadata{
		BucketName: "test-bucket",
	})

	resp, err := executor.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestEncodePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Empty path",
			input: "",
			want:  "/",
		},
		{
			name:  "Simple path",
			input: "folder/object.txt",
			want:  "folder/object.txt",
		},
		{
			name:  "Path with spaces",
			input: "folder/my object.txt",
			want:  "folder/my%20object.txt",
		},
		{
			name:  "Path with special characters",
			input: "folder/object+test.txt",
			want:  "folder/object%2Btest.txt",
		},
		{
			name:  "Path with Chinese characters",
			input: "文件夹/对象.txt",
			want:  "%E6%96%87%E4%BB%B6%E5%A4%B9/%E5%AF%B9%E8%B1%A1.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodePath(tt.input)
			if got != tt.want {
				t.Errorf("encodePath() = %s, want %s", got, tt.want)
			}
		})
	}
}

// testNetError 模拟网络错误
type testNetError struct {
	msg     string
	temp    bool
	timeout bool
}

func (e *testNetError) Error() string   { return e.msg }
func (e *testNetError) Temporary() bool { return e.temp }
func (e *testNetError) Timeout() bool   { return e.timeout }

func BenchmarkMakeTargetURL(b *testing.B) {
	endpointURL, _ := url.Parse("https://s3.amazonaws.com")
	executor := &Executor{
		endpointURL:  endpointURL,
		bucketLookup: int(types.BucketLookupPath),
	}

	for i := 0; i < b.N; i++ {
		_, _ = executor.makeTargetURL("my-bucket", "my-object.txt", "us-east-1", nil)
	}
}

func BenchmarkIsVirtualHostStyleRequest(b *testing.B) {
	endpointURL, _ := url.Parse("https://s3.amazonaws.com")
	executor := &Executor{
		endpointURL:  endpointURL,
		bucketLookup: int(types.BucketLookupAuto),
	}

	for i := 0; i < b.N; i++ {
		_ = executor.isVirtualHostStyleRequest("my-bucket")
	}
}
