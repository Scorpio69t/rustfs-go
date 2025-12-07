// Package core internal/core/executor.go
package core

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
)

// Executor 请求执行器
type Executor struct {
	// HTTP 客户端
	httpClient *http.Client

	// 端点
	endpointURL *url.URL

	// 凭证
	credentials *credentials.Credentials

	// 区域
	region string

	// 是否使用 HTTPS
	secure bool

	// 签名类型
	signerType credentials.SignatureType

	// 桶查找方式
	bucketLookup int

	// 最大重试次数
	maxRetries int

	// 位置缓存
	locationCache LocationCache

	// 调试选项
	traceEnabled bool
	traceOutput  io.Writer
}

// ExecutorConfig 执行器配置
type ExecutorConfig struct {
	HTTPClient    *http.Client
	EndpointURL   *url.URL
	Credentials   *credentials.Credentials
	Region        string
	Secure        bool
	BucketLookup  int
	MaxRetries    int
	LocationCache LocationCache
}

// NewExecutor 创建新的执行器
func NewExecutor(config ExecutorConfig) *Executor {
	maxRetries := config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 10
	}

	return &Executor{
		httpClient:    config.HTTPClient,
		endpointURL:   config.EndpointURL,
		credentials:   config.Credentials,
		region:        config.Region,
		secure:        config.Secure,
		bucketLookup:  config.BucketLookup,
		maxRetries:    maxRetries,
		locationCache: config.LocationCache,
	}
}

// Execute 执行请求
func (e *Executor) Execute(ctx context.Context, req *Request) (*http.Response, error) {
	var (
		resp    *http.Response
		err     error
		httpReq *http.Request
	)

	// 重试循环
	for attempt := 0; attempt < e.maxRetries; attempt++ {
		// 检查上下文
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// 构建 HTTP 请求
		httpReq, err = e.buildHTTPRequest(ctx, req)
		if err != nil {
			return nil, err
		}

		// 执行请求
		resp, err = e.httpClient.Do(httpReq)
		if err != nil {
			if e.shouldRetry(err, attempt) {
				e.waitForRetry(ctx, attempt)
				continue
			}
			return nil, err
		}

		// 检查响应
		if e.isSuccessStatus(resp.StatusCode, req.metadata.Expect200OKWithError) {
			return resp, nil
		}

		// 检查是否需要重试
		if e.shouldRetryResponse(resp, attempt) {
			closeResponse(resp)
			e.waitForRetry(ctx, attempt)
			continue
		}

		// 返回错误响应
		return resp, nil
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// buildHTTPRequest 构建 HTTP 请求
func (e *Executor) buildHTTPRequest(ctx context.Context, req *Request) (*http.Request, error) {
	meta := req.Metadata()

	// 获取桶位置
	location := meta.BucketLocation
	if location == "" && meta.BucketName != "" {
		location = e.getBucketLocation(ctx, meta.BucketName)
	}
	if location == "" {
		location = e.region
	}

	// 构建 URL
	targetURL, err := e.makeTargetURL(meta.BucketName, meta.ObjectName, location, meta.QueryValues)
	if err != nil {
		return nil, err
	}

	// 创建请求
	httpReq, err := http.NewRequestWithContext(ctx, req.Method(), targetURL.String(), meta.ContentBody)
	if err != nil {
		return nil, err
	}

	// 设置头部
	for k, v := range meta.CustomHeader {
		httpReq.Header[k] = v
	}

	// 设置 Content-Length
	httpReq.ContentLength = meta.ContentLength

	// 签名请求
	if err := e.signRequest(httpReq, meta, location); err != nil {
		return nil, err
	}

	return httpReq, nil
}

// makeTargetURL 构建目标 URL
func (e *Executor) makeTargetURL(bucketName, objectName, location string, queryValues url.Values) (*url.URL, error) {
	// TODO: 实现 URL 构建逻辑
	// 根据 bucketLookup 决定使用路径风格还是虚拟主机风格
	return nil, nil
}

// signRequest 签名请求
func (e *Executor) signRequest(req *http.Request, meta RequestMetadata, location string) error {
	// TODO: 实现签名逻辑
	return nil
}

// getBucketLocation 获取桶位置
func (e *Executor) getBucketLocation(ctx context.Context, bucketName string) string {
	if e.locationCache != nil {
		if loc, ok := e.locationCache.Get(bucketName); ok {
			return loc
		}
	}
	return e.region
}

// shouldRetry 判断是否应该重试
func (e *Executor) shouldRetry(err error, attempt int) bool {
	if attempt >= e.maxRetries-1 {
		return false
	}
	// TODO: 检查网络错误等
	return false
}

// shouldRetryResponse 判断响应是否应该重试
func (e *Executor) shouldRetryResponse(resp *http.Response, attempt int) bool {
	if attempt >= e.maxRetries-1 {
		return false
	}
	// 5xx 错误可重试
	if resp.StatusCode >= 500 {
		return true
	}
	// 429 Too Many Requests
	if resp.StatusCode == 429 {
		return true
	}
	return false
}

// waitForRetry 等待重试
func (e *Executor) waitForRetry(ctx context.Context, attempt int) {
	// 指数退避
	delay := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
	if delay > 10*time.Second {
		delay = 10 * time.Second
	}

	select {
	case <-ctx.Done():
	case <-time.After(delay):
	}
}

// isSuccessStatus 判断是否为成功状态
func (e *Executor) isSuccessStatus(statusCode int, expect200OKWithError bool) bool {
	if expect200OKWithError {
		return false // 需要检查响应体
	}
	return statusCode >= 200 && statusCode < 300
}

// LocationCache 位置缓存接口
type LocationCache interface {
	Get(bucketName string) (string, bool)
	Set(bucketName, location string)
	Delete(bucketName string)
}

// closeResponse 关闭响应
func closeResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}
