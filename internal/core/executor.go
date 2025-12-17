// Package core internal/core/executor.go
package core

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Scorpio69t/rustfs-go/internal/signer"
	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/types"
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

	// 设置 Content-SHA256 头（AWS Signature V4 必需）
	if meta.ContentSHA256Hex != "" {
		httpReq.Header.Set("X-Amz-Content-Sha256", meta.ContentSHA256Hex)
	}

	// 签名请求
	if err := e.signRequest(httpReq, meta, location); err != nil {
		return nil, err
	}

	return httpReq, nil
}

// makeTargetURL 构建目标 URL
func (e *Executor) makeTargetURL(bucketName, objectName, location string, queryValues url.Values) (*url.URL, error) {
	host := e.endpointURL.Host
	scheme := e.endpointURL.Scheme

	// 处理端口：去除 80 (http) 和 443 (https)
	// 原因：浏览器和 curl 会自动移除这些端口，导致预签名 URL 签名不匹配
	if h, p, err := net.SplitHostPort(host); err == nil {
		if (scheme == "http" && p == "80") || (scheme == "https" && p == "443") {
			host = h
			// 如果是 IPv6 地址，需要加方括号
			if ip := net.ParseIP(h); ip != nil && ip.To4() == nil {
				host = "[" + h + "]"
			}
		}
	}

	urlStr := scheme + "://" + host + "/"

	// 如果有桶名，构建完整 URL
	if bucketName != "" {
		// 判断是否使用虚拟主机风格
		isVirtualHost := e.isVirtualHostStyleRequest(bucketName)

		if isVirtualHost {
			// 虚拟主机风格: http://bucket.host/object
			urlStr = scheme + "://" + bucketName + "." + host + "/"
			if objectName != "" {
				urlStr += encodePath(objectName)
			}
		} else {
			// 路径风格: http://host/bucket/object
			urlStr = urlStr + bucketName + "/"
			if objectName != "" {
				urlStr += encodePath(objectName)
			}
		}
	}

	// 添加查询参数
	if len(queryValues) > 0 {
		urlStr = urlStr + "?" + queryEncode(queryValues)
	}

	return url.Parse(urlStr)
}

// isVirtualHostStyleRequest 判断是否使用虚拟主机风格
func (e *Executor) isVirtualHostStyleRequest(bucketName string) bool {
	if bucketName == "" {
		return false
	}

	lookup := types.BucketLookupType(e.bucketLookup)

	switch lookup {
	case types.BucketLookupDNS:
		return true
	case types.BucketLookupPath:
		return false
	case types.BucketLookupAuto:
		// 自动检测：检查桶名是否符合 DNS 命名规范
		return isValidVirtualHostBucket(bucketName, e.endpointURL.Scheme == "https")
	}

	return false
}

// isValidVirtualHostBucket 检查桶名是否符合虚拟主机 DNS 命名规范
func isValidVirtualHostBucket(bucketName string, https bool) bool {
	if strings.Contains(bucketName, ".") {
		// 包含点的桶名在 HTTPS 下会导致证书不匹配
		if https {
			return false
		}
	}
	// 检查桶名长度（3-63 字符）
	if len(bucketName) < 3 || len(bucketName) > 63 {
		return false
	}
	// 检查是否为 IP 地址格式
	if net.ParseIP(bucketName) != nil {
		return false
	}
	return true
}

// encodePath URL 编码路径（保留 /）
func encodePath(pathName string) string {
	if pathName == "" {
		return "/"
	}

	// S3 要求保留路径中的斜杠，但编码其他特殊字符（包括 +）
	var encodedPathname strings.Builder
	for _, segment := range strings.Split(pathName, "/") {
		if encodedPathname.Len() > 0 {
			encodedPathname.WriteByte('/')
		}
		// url.PathEscape 不会编码 +，需要手动处理
		encoded := url.PathEscape(segment)
		encoded = strings.ReplaceAll(encoded, "+", "%2B")
		encodedPathname.WriteString(encoded)
	}

	result := encodedPathname.String()
	if result == "" {
		return "/"
	}
	return result
}

// queryEncode 编码查询参数
func queryEncode(v url.Values) string {
	if v == nil {
		return ""
	}
	// url.Values.Encode() 已经按字典序排序并编码
	return v.Encode()
}

// signRequest 签名请求
func (e *Executor) signRequest(req *http.Request, meta RequestMetadata, location string) error {
	// 获取凭证
	if e.credentials == nil {
		return nil // 匿名请求
	}

	creds, err := e.credentials.Get()
	if err != nil {
		return err
	}

	// 如果是匿名凭证，不签名
	if creds.SignerType == credentials.SignatureAnonymous {
		return nil
	}

	// 使用区域（优先使用桶位置）
	region := location
	if region == "" {
		region = e.region
	}

	// 如果是预签名请求
	if meta.PresignURL {
		expires := time.Duration(meta.Expires) * time.Second
		sn := signer.NewSigner(convertSignerType(creds.SignerType))
		sn.Presign(req, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, region, expires)
		return nil
	}

	// 普通请求签名
	sn := signer.NewSigner(convertSignerType(creds.SignerType))
	sn.Sign(req, creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, region)
	return nil
}

// convertSignerType 转换签名类型
func convertSignerType(st credentials.SignatureType) signer.SignerType {
	switch st {
	case credentials.SignatureV2:
		return signer.SignerV2
	case credentials.SignatureAnonymous:
		return signer.SignerAnonymous
	default:
		return signer.SignerV4
	}
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

	// 检查错误类型
	if err == nil {
		return false
	}

	// 检查是否为网络错误
	errStr := err.Error()

	// 连接被拒绝、超时、临时错误等可重试
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "TLS handshake timeout") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "net/http: request canceled") ||
		strings.Contains(errStr, "context deadline exceeded") {
		return true
	}

	// 检查 url.Error
	if urlErr, ok := err.(*url.Error); ok {
		if urlErr.Temporary() || urlErr.Timeout() {
			return true
		}
		// 递归检查内部错误
		return e.shouldRetry(urlErr.Err, attempt)
	}

	// 检查 net.Error
	if netErr, ok := err.(net.Error); ok {
		return netErr.Temporary() || netErr.Timeout()
	}

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
