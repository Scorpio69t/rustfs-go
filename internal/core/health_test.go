package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Scorpio69t/rustfs-go/internal/cache"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

func TestHealthCheck(t *testing.T) {
	tests := []struct {
		name              string
		statusCode        int  // HEAD 请求返回的状态码
		expectedFinalCode int  // 期望的最终状态码（可能经过 GET 重试）
		bucketName        string
		expectedHealth    bool
		delay             time.Duration
	}{
		{
			name:              "Healthy endpoint - 200 OK",
			statusCode:        http.StatusOK,
			expectedFinalCode: http.StatusOK,
			expectedHealth:    true,
		},
		{
			name:              "Healthy endpoint - 204 No Content",
			statusCode:        http.StatusNoContent,
			expectedFinalCode: http.StatusNoContent,
			expectedHealth:    true,
		},
		{
			name:              "Root endpoint - 403 Forbidden (still healthy)",
			statusCode:        http.StatusForbidden,
			expectedFinalCode: http.StatusForbidden,
			expectedHealth:    true,
		},
		{
			name:              "Bucket endpoint - 403 Forbidden (healthy, exists but needs auth)",
			statusCode:        http.StatusForbidden,
			expectedFinalCode: http.StatusForbidden,
			bucketName:        "test-bucket",
			expectedHealth:    true, // 403 表示存储桶存在，只是需要认证
		},
		{
			name:              "Bucket endpoint - 404 Not Found (bucket does not exist)",
			statusCode:        http.StatusNotFound,
			expectedFinalCode: http.StatusNotFound,
			bucketName:        "non-existent-bucket",
			expectedHealth:    false, // 404 表示存储桶不存在
		},
		{
			name:              "Root endpoint - 404 Not Found",
			statusCode:        http.StatusNotFound,
			expectedFinalCode: http.StatusNotFound,
			expectedHealth:    false,
		},
		{
			name:              "Unhealthy endpoint - 500 Internal Server Error",
			statusCode:        http.StatusInternalServerError,
			expectedFinalCode: http.StatusInternalServerError,
			expectedHealth:    false,
		},
		{
			name:              "Slow endpoint",
			statusCode:        http.StatusOK,
			expectedFinalCode: http.StatusOK,
			expectedHealth:    true,
			delay:             100 * time.Millisecond,
		},
		{
			name:              "Root endpoint - 501 Not Implemented (fallback to GET)",
			statusCode:        http.StatusNotImplemented, // HEAD 返回 501
			expectedFinalCode: http.StatusOK,              // GET 返回 200
			expectedHealth:    true,                       // 应该回退到 GET 请求并成功
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟服务器
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.delay > 0 {
					time.Sleep(tt.delay)
				}
				// 如果配置的状态码是 501，但请求是 GET，则返回 200
				if tt.statusCode == http.StatusNotImplemented && r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(tt.statusCode)
				}
			}))
			defer server.Close()

			// 解析服务器 URL
			endpointURL, _ := url.Parse(server.URL)

			// 创建执行器
			executor := &Executor{
				httpClient:    server.Client(),
				endpointURL:   endpointURL,
				region:        "us-east-1",
				credentials:   credentials.NewStaticV4("test", "test", ""),
				signerType:    credentials.SignatureV4,
				bucketLookup:  0,
				maxRetries:    3,
				locationCache: &cache.LocationCache{},
			}

			// 执行健康检查
			opts := &HealthCheckOptions{
				Timeout:    2 * time.Second,
				BucketName: tt.bucketName,
				Context:    context.Background(),
			}

			result := executor.HealthCheck(opts)

			// 验证结果
			if result.Healthy != tt.expectedHealth {
				t.Errorf("Expected healthy=%v, got %v", tt.expectedHealth, result.Healthy)
			}

			expectedCode := tt.expectedFinalCode
			if expectedCode == 0 {
				expectedCode = tt.statusCode
			}
			if result.StatusCode != expectedCode {
				t.Errorf("Expected status code %d, got %d", expectedCode, result.StatusCode)
			}

			if result.Endpoint == "" {
				t.Error("Expected endpoint to be set")
			}

			if result.CheckedAt.IsZero() {
				t.Error("Expected CheckedAt to be set")
			}

			// ResponseTime 可能非常短，接近 0，所以我们只记录而不报错
			if result.ResponseTime == 0 {
				t.Logf("Warning: ResponseTime is 0 (may be very fast)")
			}

			t.Logf("Result: %s", result.String())
		})
	}
}

func TestHealthCheckTimeout(t *testing.T) {
	// 创建一个慢响应的服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	endpointURL, _ := url.Parse(server.URL)

	executor := &Executor{
		httpClient:    server.Client(),
		endpointURL:   endpointURL,
		region:        "us-east-1",
		credentials:   credentials.NewStaticV4("test", "test", ""),
		signerType:    credentials.SignatureV4,
		bucketLookup:  0,
		maxRetries:    3,
		locationCache: &cache.LocationCache{},
	}

	// 设置较短的超时
	opts := &HealthCheckOptions{
		Timeout: 500 * time.Millisecond,
		Context: context.Background(),
	}

	result := executor.HealthCheck(opts)

	// 应该超时，不健康
	if result.Healthy {
		t.Error("Expected unhealthy due to timeout")
	}

	if result.Error == nil {
		t.Error("Expected error due to timeout")
	}

	t.Logf("Timeout result: %s", result.String())
}

func TestHealthCheckWithRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// 前两次返回错误
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			// 第三次返回成功
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	endpointURL, _ := url.Parse(server.URL)

	executor := &Executor{
		httpClient:    server.Client(),
		endpointURL:   endpointURL,
		region:        "us-east-1",
		credentials:   credentials.NewStaticV4("test", "test", ""),
		signerType:    credentials.SignatureV4,
		bucketLookup:  0,
		maxRetries:    3,
		locationCache: &cache.LocationCache{},
	}

	opts := &HealthCheckOptions{
		Timeout: 2 * time.Second,
		Context: context.Background(),
	}

	result := executor.HealthCheckWithRetry(opts, 3)

	// 应该在第三次成功
	if !result.Healthy {
		t.Errorf("Expected healthy after retry, got: %s", result.String())
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}

	t.Logf("Retry result: %s", result.String())
}

func TestHealthCheckDefaultOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	endpointURL, _ := url.Parse(server.URL)

	executor := &Executor{
		httpClient:    server.Client(),
		endpointURL:   endpointURL,
		region:        "us-east-1",
		credentials:   credentials.NewStaticV4("test", "test", ""),
		signerType:    credentials.SignatureV4,
		bucketLookup:  0,
		maxRetries:    3,
		locationCache: &cache.LocationCache{},
	}

	// 使用 nil 选项（应该使用默认值）
	result := executor.HealthCheck(nil)

	if !result.Healthy {
		t.Error("Expected healthy with default options")
	}

	t.Logf("Default options result: %s", result.String())
}

func TestHealthCheckResultString(t *testing.T) {
	tests := []struct {
		name   string
		result *HealthCheckResult
	}{
		{
			name: "Healthy result",
			result: &HealthCheckResult{
				Healthy:      true,
				Endpoint:     "http://localhost:9000",
				Region:       "us-east-1",
				ResponseTime: 100 * time.Millisecond,
				StatusCode:   200,
			},
		},
		{
			name: "Unhealthy result",
			result: &HealthCheckResult{
				Healthy:      false,
				Endpoint:     "http://localhost:9000",
				Region:       "us-east-1",
				ResponseTime: 50 * time.Millisecond,
				StatusCode:   500,
				Error:        http.ErrServerClosed,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.result.String()
			if str == "" {
				t.Error("Expected non-empty string representation")
			}
			t.Logf("String representation: %s", str)
		})
	}
}

func BenchmarkHealthCheck(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	endpointURL, _ := url.Parse(server.URL)

	executor := &Executor{
		httpClient:    server.Client(),
		endpointURL:   endpointURL,
		region:        "us-east-1",
		credentials:   credentials.NewStaticV4("test", "test", ""),
		signerType:    credentials.SignatureV4,
		bucketLookup:  0,
		maxRetries:    3,
		locationCache: &cache.LocationCache{},
	}

	opts := &HealthCheckOptions{
		Timeout: 2 * time.Second,
		Context: context.Background(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.HealthCheck(opts)
	}
}
