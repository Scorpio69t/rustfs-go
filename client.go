// Package rustfs client.go - RustFS Go SDK 客户端入口
package rustfs

import (
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/Scorpio69t/rustfs-go/bucket"
	"github.com/Scorpio69t/rustfs-go/internal/cache"
	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/internal/transport"
	"github.com/Scorpio69t/rustfs-go/object"
	"github.com/Scorpio69t/rustfs-go/types"
	"golang.org/x/net/publicsuffix"
)

// Client RustFS 客户端
type Client struct {
	// 核心组件
	executor      *core.Executor
	locationCache *cache.LocationCache

	// 服务模块
	bucketService bucket.Service
	objectService object.Service

	// 客户端信息
	endpointURL *url.URL
	httpClient  *http.Client
	secure      bool
	region      string

	// 应用信息
	appInfo struct {
		appName    string
		appVersion string
	}
}

// New 创建新的 RustFS 客户端
//
// Parameters:
//   - endpoint: RustFS 服务器地址 (e.g., "localhost:9000", "rustfs.example.com")
//   - opts: 客户端配置选项
//
// Returns:
//   - *Client: 客户端实例
//   - error: 错误信息
//
// Example:
//
//	client, err := rustfs.New("localhost:9000", &rustfs.Options{
//	    Credentials: credentials.NewStaticV4("access-key", "secret-key", ""),
//	    Secure:      false,
//	})
func New(endpoint string, opts *Options) (*Client, error) {
	// 验证选项
	if err := opts.validate(); err != nil {
		return nil, err
	}

	// 设置默认值
	opts.setDefaults()

	// 解析 endpoint URL
	endpointURL, err := parseEndpointURL(endpoint, opts.Secure)
	if err != nil {
		return nil, err
	}

	// 如果 BucketLookup 是 Auto，且 endpoint 是 IP 地址，使用 Path 风格
	if opts.BucketLookup == types.BucketLookupAuto && isIPAddress(endpointURL.Host) {
		opts.BucketLookup = types.BucketLookupPath
	}

	// 创建 HTTP Transport
	var httpTransport http.RoundTripper
	if opts.Transport != nil {
		httpTransport = opts.Transport
	} else {
		httpTransport = transport.NewTransport(transport.TransportOptions{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90,
			EnableCompression:   false,
		})
	}

	// 创建 Cookie Jar
	jar, err := cookiejar.New(&cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	})
	if err != nil {
		return nil, err
	}

	// 创建 HTTP 客户端
	httpClient := &http.Client{
		Jar:       jar,
		Transport: httpTransport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 确定区域
	region := opts.Region
	if region == "" {
		region = detectRegion(endpointURL, opts.CustomRegionViaURL)
	}

	// 创建位置缓存
	locationCache := cache.NewLocationCache(0) // 0 = 无超时

	// 创建核心执行器
	executor := core.NewExecutor(core.ExecutorConfig{
		HTTPClient:   httpClient,
		EndpointURL:  endpointURL,
		Credentials:  opts.Credentials,
		Region:       region,
		BucketLookup: int(opts.BucketLookup),
		MaxRetries:   opts.MaxRetries,
	})

	// 创建服务实例
	bucketService := bucket.NewService(executor, locationCache)
	objectService := object.NewService(executor, locationCache)

	// 创建客户端
	client := &Client{
		executor:      executor,
		locationCache: locationCache,
		bucketService: bucketService,
		objectService: objectService,
		endpointURL:   endpointURL,
		httpClient:    httpClient,
		secure:        opts.Secure,
		region:        region,
	}

	return client, nil
}

// Bucket 返回 Bucket 服务接口
//
// Example:
//
//	err := client.Bucket().Create(ctx, "my-bucket")
func (c *Client) Bucket() bucket.Service {
	return c.bucketService
}

// Object 返回 Object 服务接口
//
// Example:
//
//	info, err := client.Object().Put(ctx, "my-bucket", "my-object", reader, size)
func (c *Client) Object() object.Service {
	return c.objectService
}

// EndpointURL 返回客户端使用的 Endpoint URL
func (c *Client) EndpointURL() *url.URL {
	endpoint := *c.endpointURL // 复制以防止修改内部状态
	return &endpoint
}

// Region 返回客户端使用的区域
func (c *Client) Region() string {
	return c.region
}

// IsSecure 返回客户端是否使用 HTTPS
func (c *Client) IsSecure() bool {
	return c.secure
}

// SetAppInfo 设置应用程序信息
//
// 这将添加到 User-Agent 头中，帮助在服务器日志中识别您的应用程序
//
// Parameters:
//   - appName: 应用程序名称
//   - appVersion: 应用程序版本
func (c *Client) SetAppInfo(appName, appVersion string) {
	if appName != "" && appVersion != "" {
		c.appInfo.appName = appName
		c.appInfo.appVersion = appVersion
	}
}

// parseEndpointURL 解析 endpoint URL
func parseEndpointURL(endpoint string, secure bool) (*url.URL, error) {
	if endpoint == "" {
		return nil, errInvalidArgument("endpoint cannot be empty")
	}

	// 如果没有协议，添加默认协议
	scheme := "http"
	if secure {
		scheme = "https"
	}

	// 检查是否已经有 scheme
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		// 没有 scheme，添加默认 scheme
		endpoint = scheme + "://" + endpoint
	}

	// 解析 URL
	endpointURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	// 验证 scheme
	if endpointURL.Scheme != "http" && endpointURL.Scheme != "https" {
		return nil, errInvalidArgument("endpoint scheme must be http or https")
	}

	return endpointURL, nil
}

// detectRegion 检测区域
func detectRegion(endpointURL *url.URL, customRegionFn func(url.URL) string) string {
	if customRegionFn != nil {
		return customRegionFn(*endpointURL)
	}

	// 默认区域
	return "us-east-1"
}

// isIPAddress 检查主机是否是 IP 地址（包括端口）
func isIPAddress(host string) bool {
	// 移除端口号
	hostOnly := host
	if colonIndex := strings.LastIndex(host, ":"); colonIndex != -1 {
		hostOnly = host[:colonIndex]
	}

	// 检查是否是 IP 地址
	return net.ParseIP(hostOnly) != nil
}

// HealthCheck 执行服务健康检查
//
// 通过发送一个简单的 HEAD 请求来验证与 RustFS 服务器的连接是否正常
//
// Parameters:
//   - opts: 健康检查选项（可以为 nil 使用默认值）
//
// Returns:
//   - *core.HealthCheckResult: 健康检查结果
//
// Example:
//
//	result := client.HealthCheck(nil)
//	if result.Healthy {
//	    fmt.Printf("服务健康，响应时间: %v\n", result.ResponseTime)
//	}
func (c *Client) HealthCheck(opts *core.HealthCheckOptions) *core.HealthCheckResult {
	return c.executor.HealthCheck(opts)
}

// HealthCheckWithRetry 执行带重试的健康检查
//
// 如果第一次检查失败，会自动重试指定次数
//
// Parameters:
//   - opts: 健康检查选项
//   - maxRetries: 最大重试次数（如果 <= 0，默认为 3）
//
// Returns:
//   - *core.HealthCheckResult: 最终的健康检查结果
//
// Example:
//
//	result := client.HealthCheckWithRetry(&core.HealthCheckOptions{
//	    Timeout: 5 * time.Second,
//	}, 3)
func (c *Client) HealthCheckWithRetry(opts *core.HealthCheckOptions, maxRetries int) *core.HealthCheckResult {
	return c.executor.HealthCheckWithRetry(opts, maxRetries)
}
