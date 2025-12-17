// Package rustfs client.go - RustFS Go SDK 客户端入口
package rustfs

import (
	"context"
	"io"
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

// --- 向后兼容的快捷方法 ---
// 以下方法保持与旧 API 的兼容性，但标记为 Deprecated

// MakeBucket 创建存储桶
//
// Deprecated: 使用 client.Bucket().Create() 代替
func (c *Client) MakeBucket(ctx context.Context, bucketName string, opts MakeBucketOptions) error {
	return c.bucketService.Create(ctx, bucketName,
		bucket.WithRegion(opts.Region),
		bucket.WithObjectLocking(opts.ObjectLocking),
	)
}

// RemoveBucket 删除存储桶
//
// Deprecated: 使用 client.Bucket().Delete() 代替
func (c *Client) RemoveBucket(ctx context.Context, bucketName string) error {
	return c.bucketService.Delete(ctx, bucketName)
}

// BucketExists 检查存储桶是否存在
//
// Deprecated: 使用 client.Bucket().Exists() 代替
func (c *Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return c.bucketService.Exists(ctx, bucketName)
}

// ListBuckets 列出所有存储桶
//
// Deprecated: 使用 client.Bucket().List() 代替
func (c *Client) ListBuckets(ctx context.Context) ([]types.BucketInfo, error) {
	return c.bucketService.List(ctx)
}

// GetBucketLocation 获取存储桶位置
//
// Deprecated: 使用 client.Bucket().GetLocation() 代替
func (c *Client) GetBucketLocation(ctx context.Context, bucketName string) (string, error) {
	return c.bucketService.GetLocation(ctx, bucketName)
}

// PutObject 上传对象
//
// Deprecated: 使用 client.Object().Put() 代替
func (c *Client) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts PutObjectOptions) (types.UploadInfo, error) {
	return c.objectService.Put(ctx, bucketName, objectName, reader, objectSize,
		convertPutOptions(opts)...,
	)
}

// GetObject 下载对象
//
// Deprecated: 使用 client.Object().Get() 代替
func (c *Client) GetObject(ctx context.Context, bucketName, objectName string, opts GetObjectOptions) (io.ReadCloser, types.ObjectInfo, error) {
	return c.objectService.Get(ctx, bucketName, objectName,
		convertGetOptions(opts)...,
	)
}

// StatObject 获取对象信息
//
// Deprecated: 使用 client.Object().Stat() 代替
func (c *Client) StatObject(ctx context.Context, bucketName, objectName string, opts StatObjectOptions) (types.ObjectInfo, error) {
	return c.objectService.Stat(ctx, bucketName, objectName,
		convertStatOptions(opts)...,
	)
}

// RemoveObject 删除对象
//
// Deprecated: 使用 client.Object().Delete() 代替
func (c *Client) RemoveObject(ctx context.Context, bucketName, objectName string, opts RemoveObjectOptions) error {
	return c.objectService.Delete(ctx, bucketName, objectName,
		convertDeleteOptions(opts)...,
	)
}

// ListObjects 列出对象
//
// Deprecated: 使用 client.Object().List() 代替
func (c *Client) ListObjects(ctx context.Context, bucketName string, opts ListObjectsOptions) <-chan types.ObjectInfo {
	return c.objectService.List(ctx, bucketName,
		convertListOptions(opts)...,
	)
}

// CopyObject 复制对象
//
// Deprecated: 使用 client.Object().Copy() 代替
func (c *Client) CopyObject(ctx context.Context, destBucket, destObject, sourceBucket, sourceObject string, opts CopyObjectOptions) (types.CopyInfo, error) {
	return c.objectService.Copy(ctx, destBucket, destObject, sourceBucket, sourceObject,
		convertCopyOptions(opts)...,
	)
}
