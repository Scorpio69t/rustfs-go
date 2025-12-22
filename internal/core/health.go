// Package core internal/core/health.go
package core

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HealthCheckResult holds health check result
type HealthCheckResult struct {
	// Whether healthy
	Healthy bool

	// Error detail
	Error error

	// Response time
	ResponseTime time.Duration

	// HTTP status code
	StatusCode int

	// Checked time
	CheckedAt time.Time

	// Endpoint
	Endpoint string

	// Region
	Region string
}

// HealthCheckOptions health check options
type HealthCheckOptions struct {
	// Timeout (default 5 seconds)
	Timeout time.Duration

	// Optional bucket name for check (default none)
	BucketName string

	// Context
	Context context.Context
}

// HealthCheck performs health check
// Sends a simple HEAD request to the endpoint to verify connectivity
func (e *Executor) HealthCheck(opts *HealthCheckOptions) *HealthCheckResult {
	result := &HealthCheckResult{
		Endpoint:  e.endpointURL.String(),
		Region:    e.region,
		CheckedAt: time.Now(),
		Healthy:   false,
	}

	// Set default options
	if opts == nil {
		opts = &HealthCheckOptions{}
	}
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Second
	}
	if opts.Context == nil {
		opts.Context = context.Background()
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(opts.Context, opts.Timeout)
	defer cancel()

	// Build request URL
	var reqURL string
	if opts.BucketName != "" {
		// Check specific bucket
		reqURL = e.endpointURL.String() + "/" + opts.BucketName
	} else {
		// Check root endpoint
		reqURL = e.endpointURL.String() + "/"
	}

	// Try HEAD request first
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, reqURL, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create health check request: %w", err)
		return result
	}

	// Record start time
	startTime := time.Now()

	// Send request
	resp, err := e.httpClient.Do(req)

	// Record response time
	result.ResponseTime = time.Since(startTime)

	if err != nil {
		result.Error = fmt.Errorf("health check request failed: %w", err)
		return result
	}

	result.StatusCode = resp.StatusCode

	// If HEAD returns 501, fallback to GET
	if resp.StatusCode == http.StatusNotImplemented {
		resp.Body.Close() // close first response

		req, err = http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			result.Error = fmt.Errorf("failed to create GET health check request: %w", err)
			return result
		}

		startTime = time.Now()
		resp, err = e.httpClient.Do(req)
		result.ResponseTime = time.Since(startTime)

		if err != nil {
			result.Error = fmt.Errorf("GET health check request failed: %w", err)
			return result
		}

		result.StatusCode = resp.StatusCode
	}

	// Ensure response body closed
	defer resp.Body.Close()

	// Determine health:
	// 2xx/3xx considered healthy
	// 403 considered healthy (reachable but requires auth)
	// 404 is unhealthy for bucket checks (bucket missing)
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.Healthy = true
	} else if resp.StatusCode == http.StatusForbidden {
		// 403 means resource exists but requires auth; treat as healthy
		result.Healthy = true
	}

	return result
}

// HealthCheckWithRetry performs health check with retries
func (e *Executor) HealthCheckWithRetry(opts *HealthCheckOptions, maxRetries int) *HealthCheckResult {
	if maxRetries <= 0 {
		maxRetries = 3
	}

	var lastResult *HealthCheckResult

	for i := 0; i < maxRetries; i++ {
		result := e.HealthCheck(opts)

		if result.Healthy {
			return result
		}

		lastResult = result

		// Wait before next retry if not last
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return lastResult
}

// String returns formatted health check result
func (r *HealthCheckResult) String() string {
	if r.Healthy {
		return fmt.Sprintf("Healthy - Endpoint: %s, Region: %s, ResponseTime: %v, StatusCode: %d",
			r.Endpoint, r.Region, r.ResponseTime, r.StatusCode)
	}
	return fmt.Sprintf("Unhealthy - Endpoint: %s, Region: %s, ResponseTime: %v, StatusCode: %d, Error: %v",
		r.Endpoint, r.Region, r.ResponseTime, r.StatusCode, r.Error)
}
