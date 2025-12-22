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
		statusCode        int // status code returned by HEAD
		expectedFinalCode int // expected final status (may be after GET retry)
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
			expectedHealth:    true, // 403 means bucket exists but needs auth
		},
		{
			name:              "Bucket endpoint - 404 Not Found (bucket does not exist)",
			statusCode:        http.StatusNotFound,
			expectedFinalCode: http.StatusNotFound,
			bucketName:        "non-existent-bucket",
			expectedHealth:    false, // 404 means bucket does not exist
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
			statusCode:        http.StatusNotImplemented, // HEAD returns 501
			expectedFinalCode: http.StatusOK,             // GET returns 200
			expectedHealth:    true,                      // should fallback to GET and succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.delay > 0 {
					time.Sleep(tt.delay)
				}
				// If configured 501 but request is GET, return 200
				if tt.statusCode == http.StatusNotImplemented && r.Method == http.MethodGet {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(tt.statusCode)
				}
			}))
			defer server.Close()

			// Parse server URL
			endpointURL, _ := url.Parse(server.URL)

			// Create executor
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

			// Perform health check
			opts := &HealthCheckOptions{
				Timeout:    2 * time.Second,
				BucketName: tt.bucketName,
				Context:    context.Background(),
			}

			result := executor.HealthCheck(opts)

			// Validate result
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

			// ResponseTime may be very short (near 0); log but don't fail
			if result.ResponseTime == 0 {
				t.Logf("Warning: ResponseTime is 0 (may be very fast)")
			}

			t.Logf("Result: %s", result.String())
		})
	}
}

func TestHealthCheckTimeout(t *testing.T) {
	// Create a slow server
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

	// Set shorter timeout
	opts := &HealthCheckOptions{
		Timeout: 500 * time.Millisecond,
		Context: context.Background(),
	}

	result := executor.HealthCheck(opts)

	// Should time out and be unhealthy
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
			// First two attempts return error
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			// Third attempt returns success
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

	// Should succeed on third attempt
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

	// Use nil options (should apply defaults)
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
