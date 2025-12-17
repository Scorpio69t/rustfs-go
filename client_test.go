// Package rustfs client_test.go
package rustfs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

const (
	YOURACCESSKEYID     = "XhJOoEKn3BM6cjD2dVmx"
	YOURSECRETACCESSKEY = "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf"
	YOURENDPOINT        = "127.0.0.1:9000"
	YOURBUCKET          = "mybucket" // 'mc mb play/mybucket' if it does not exist.
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		opts     *Options
		wantErr  bool
	}{
		{
			name:     "Valid client creation",
			endpoint: "127.0.0.1:9000",
			opts: &Options{
				Credentials: credentials.NewStaticV4("XhJOoEKn3BM6cjD2dVmx", "yXKl1p5FNjgWdqHzYV8s3LTuoxAEBwmb67DnchRf", ""),
				Secure:      false,
			},
			wantErr: false,
		},
		{
			name:     "Valid client with HTTPS",
			endpoint: "rustfs.example.com",
			opts: &Options{
				Credentials: credentials.NewStaticV4("access-key", "secret-key", ""),
				Secure:      true,
			},
			wantErr: false,
		},
		{
			name:     "Empty endpoint",
			endpoint: "",
			opts: &Options{
				Credentials: credentials.NewStaticV4("access-key", "secret-key", ""),
			},
			wantErr: true,
		},
		{
			name:     "Nil credentials",
			endpoint: "127.0.0.1:9000",
			opts: &Options{
				Credentials: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.endpoint, tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("New() returned nil client")
			}
		})
	}
}

func TestClientMethods(t *testing.T) {
	client, err := New("127.0.0.1:9000", &Options{
		Credentials: credentials.NewStaticV4("access-key", "secret-key", ""),
		Secure:      false,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("Bucket service", func(t *testing.T) {
		if client.Bucket() == nil {
			t.Error("Bucket() returned nil")
		}
	})

	t.Run("Object service", func(t *testing.T) {
		if client.Object() == nil {
			t.Error("Object() returned nil")
		}
	})

	t.Run("EndpointURL", func(t *testing.T) {
		url := client.EndpointURL()
		if url == nil {
			t.Error("EndpointURL() returned nil")
		}
		if url.Host != "127.0.0.1:9000" {
			t.Errorf("EndpointURL() host = %s, want 127.0.0.1:9000", url.Host)
		}
	})

	t.Run("Region", func(t *testing.T) {
		region := client.Region()
		if region == "" {
			t.Error("Region() returned empty string")
		}
	})

	t.Run("IsSecure", func(t *testing.T) {
		if client.IsSecure() {
			t.Error("IsSecure() = true, want false")
		}
	})

	t.Run("SetAppInfo", func(t *testing.T) {
		client.SetAppInfo("test-app", "1.0.0")
		if client.appInfo.appName != "test-app" {
			t.Errorf("SetAppInfo() appName = %s, want test-app", client.appInfo.appName)
		}
		if client.appInfo.appVersion != "1.0.0" {
			t.Errorf("SetAppInfo() appVersion = %s, want 1.0.0", client.appInfo.appVersion)
		}
	})
}

func TestBackwardCompatibility(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "test-bucket"):
			// MakeBucket
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "test-bucket"):
			// RemoveBucket
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodHead && strings.Contains(r.URL.Path, "test-bucket"):
			// BucketExists
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodGet && r.URL.Path == "/":
			// ListBuckets
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<ListAllMyBucketsResult>
    <Buckets>
        <Bucket>
            <Name>test-bucket</Name>
            <CreationDate>2023-01-01T00:00:00Z</CreationDate>
        </Bucket>
    </Buckets>
</ListAllMyBucketsResult>`))
		case r.Method == http.MethodGet && strings.Contains(r.URL.Query().Get("location"), ""):
			// GetBucketLocation
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<LocationConstraint>us-east-1</LocationConstraint>`))
		default:
			w.WriteHeader(http.StatusNotImplemented)
		}
	}))
	defer server.Close()

	// 创建客户端（使用完整的 URL）
	client, err := New(server.URL, &Options{
		Credentials: credentials.NewStaticV4("access-key", "secret-key", ""),
		Secure:      false,
		Transport:   server.Client().Transport,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	t.Run("MakeBucket", func(t *testing.T) {
		err := client.MakeBucket(ctx, "test-bucket", MakeBucketOptions{
			Region: "us-east-1",
		})
		if err != nil {
			t.Errorf("MakeBucket() error = %v", err)
		}
	})

	t.Run("BucketExists", func(t *testing.T) {
		exists, err := client.BucketExists(ctx, "test-bucket")
		if err != nil {
			t.Errorf("BucketExists() error = %v", err)
		}
		if !exists {
			t.Error("BucketExists() = false, want true")
		}
	})

	t.Run("ListBuckets", func(t *testing.T) {
		buckets, err := client.ListBuckets(ctx)
		if err != nil {
			t.Errorf("ListBuckets() error = %v", err)
		}
		if len(buckets) == 0 {
			t.Error("ListBuckets() returned empty list")
		}
	})

	t.Run("RemoveBucket", func(t *testing.T) {
		err := client.RemoveBucket(ctx, "test-bucket")
		if err != nil {
			t.Errorf("RemoveBucket() error = %v", err)
		}
	})
}

func TestParseEndpointURL(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		secure   bool
		wantHost string
		wantErr  bool
	}{
		{
			name:     "Simple endpoint",
			endpoint: "127.0.0.1:9000",
			secure:   false,
			wantHost: "127.0.0.1:9000",
			wantErr:  false,
		},
		{
			name:     "HTTPS endpoint",
			endpoint: "rustfs.example.com",
			secure:   true,
			wantHost: "rustfs.example.com",
			wantErr:  false,
		},
		{
			name:     "Endpoint with scheme",
			endpoint: "http://127.0.0.1:9000",
			secure:   false,
			wantHost: "127.0.0.1:9000",
			wantErr:  false,
		},
		{
			name:     "Empty endpoint",
			endpoint: "",
			secure:   false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := parseEndpointURL(tt.endpoint, tt.secure)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEndpointURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && url.Host != tt.wantHost {
				t.Errorf("parseEndpointURL() host = %s, want %s", url.Host, tt.wantHost)
			}
		})
	}
}

func BenchmarkNew(b *testing.B) {
	opts := &Options{
		Credentials: credentials.NewStaticV4("access-key", "secret-key", ""),
		Secure:      false,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = New("127.0.0.1:9000", opts)
	}
}

func BenchmarkClientBucket(b *testing.B) {
	client, _ := New("127.0.0.1:9000", &Options{
		Credentials: credentials.NewStaticV4("access-key", "secret-key", ""),
		Secure:      false,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Bucket()
	}
}

func BenchmarkClientObject(b *testing.B) {
	client, _ := New("127.0.0.1:9000", &Options{
		Credentials: credentials.NewStaticV4("access-key", "secret-key", ""),
		Secure:      false,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Object()
	}
}
