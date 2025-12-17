// Package core internal/core/health.go
package core

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	// 是否健康
	Healthy bool

	// 错误信息
	Error error

	// 响应时间
	ResponseTime time.Duration

	// HTTP 状态码
	StatusCode int

	// 检查时间
	CheckedAt time.Time

	// 端点
	Endpoint string

	// 区域
	Region string
}

// HealthCheckOptions 健康检查选项
type HealthCheckOptions struct {
	// 超时时间（默认 5 秒）
	Timeout time.Duration

	// 自定义存储桶名（用于检查，默认不使用存储桶）
	BucketName string

	// 上下文
	Context context.Context
}

// HealthCheck 执行健康检查
// 通过发送一个简单的 HEAD 请求到服务端点来验证连接
func (e *Executor) HealthCheck(opts *HealthCheckOptions) *HealthCheckResult {
	result := &HealthCheckResult{
		Endpoint:  e.endpointURL.String(),
		Region:    e.region,
		CheckedAt: time.Now(),
		Healthy:   false,
	}

	// 设置默认选项
	if opts == nil {
		opts = &HealthCheckOptions{}
	}
	if opts.Timeout == 0 {
		opts.Timeout = 5 * time.Second
	}
	if opts.Context == nil {
		opts.Context = context.Background()
	}

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(opts.Context, opts.Timeout)
	defer cancel()

	// 构建请求
	var reqURL string
	if opts.BucketName != "" {
		// 检查特定存储桶
		reqURL = e.endpointURL.String() + "/" + opts.BucketName
	} else {
		// 检查根端点
		reqURL = e.endpointURL.String() + "/"
	}

	// 先尝试 HEAD 请求
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, reqURL, nil)
	if err != nil {
		result.Error = fmt.Errorf("failed to create health check request: %w", err)
		return result
	}

	// 记录开始时间
	startTime := time.Now()

	// 发送请求
	resp, err := e.httpClient.Do(req)

	// 记录响应时间
	result.ResponseTime = time.Since(startTime)

	if err != nil {
		result.Error = fmt.Errorf("health check request failed: %w", err)
		return result
	}

	result.StatusCode = resp.StatusCode

	// 如果 HEAD 请求返回 501 (Not Implemented)，尝试使用 GET 请求
	if resp.StatusCode == http.StatusNotImplemented {
		resp.Body.Close() // 关闭第一个响应

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

	// 确保响应体被关闭
	defer resp.Body.Close()

	// 判断是否健康
	// 2xx 和 3xx 状态码认为是健康的
	// 403 也认为是健康的（服务/存储桶可达，只是需要认证）
	// 404 对于存储桶检查是不健康的（存储桶不存在）
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.Healthy = true
	} else if resp.StatusCode == http.StatusForbidden {
		// 403 表示资源存在但需要认证，认为是健康的
		result.Healthy = true
	}

	return result
}

// HealthCheckWithRetry 执行带重试的健康检查
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

		// 如果不是最后一次重试，等待一段时间
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	return lastResult
}

// String 返回健康检查结果的字符串表示
func (r *HealthCheckResult) String() string {
	if r.Healthy {
		return fmt.Sprintf("Healthy - Endpoint: %s, Region: %s, ResponseTime: %v, StatusCode: %d",
			r.Endpoint, r.Region, r.ResponseTime, r.StatusCode)
	}
	return fmt.Sprintf("Unhealthy - Endpoint: %s, Region: %s, ResponseTime: %v, StatusCode: %d, Error: %v",
		r.Endpoint, r.Region, r.ResponseTime, r.StatusCode, r.Error)
}
