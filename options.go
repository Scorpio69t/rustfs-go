// Package rustfs options.go
package rustfs

import (
	"net/http"
	"net/http/httptrace"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/pkg/credentials"
	"github.com/Scorpio69t/rustfs-go/types"
)

// Options 客户端配置选项
type Options struct {
	// Credentials 凭证提供者
	// 必需，用于签名请求
	Credentials *credentials.Credentials

	// Secure 是否使用 HTTPS
	// 默认: false
	Secure bool

	// Region 区域
	// 如果不设置，将自动检测
	Region string

	// Transport 自定义 HTTP 传输
	// 如果不设置，使用默认传输
	Transport http.RoundTripper

	// Trace HTTP 追踪客户端
	Trace *httptrace.ClientTrace

	// BucketLookup 桶查找类型
	// 默认: BucketLookupAuto
	BucketLookup types.BucketLookupType

	// CustomRegionViaURL 自定义区域查找函数
	CustomRegionViaURL func(u url.URL) string

	// BucketLookupViaURL 自定义桶查找函数
	BucketLookupViaURL func(u url.URL, bucketName string) types.BucketLookupType

	// TrailingHeaders 启用尾部头（用于流式上传）
	// 需要服务器支持
	TrailingHeaders bool

	// MaxRetries 最大重试次数
	// 默认: 10，设置为 1 禁用重试
	MaxRetries int
}

// validate 验证选项
func (o *Options) validate() error {
	if o == nil {
		return errInvalidArgument("options cannot be nil")
	}
	if o.Credentials == nil {
		return errInvalidArgument("credentials are required")
	}
	return nil
}

// setDefaults 设置默认值
func (o *Options) setDefaults() {
	if o.MaxRetries <= 0 {
		o.MaxRetries = 10
	}
	if o.BucketLookup == 0 {
		o.BucketLookup = types.BucketLookupAuto
	}
}

// errInvalidArgument 创建无效参数错误
func errInvalidArgument(message string) error {
	return &invalidArgumentError{message: message}
}

// invalidArgumentError 无效参数错误类型
type invalidArgumentError struct {
	message string
}

// Error 返回错误消息
func (e *invalidArgumentError) Error() string {
	return e.message
}
