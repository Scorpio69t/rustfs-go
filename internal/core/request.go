// Package core internal/core/request.go
package core

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

// RequestMetadata 请求元数据
type RequestMetadata struct {
	// 桶和对象
	BucketName string
	ObjectName string

	// 查询参数
	QueryValues url.Values

	// 请求头
	CustomHeader http.Header

	// 请求体
	ContentBody   io.Reader
	ContentLength int64

	// 内容校验
	ContentMD5Base64 string
	ContentSHA256Hex string

	// 签名选项
	StreamSHA256 bool
	PresignURL   bool
	Expires      int64

	// 预签名额外头
	ExtraPresignHeader http.Header

	// 位置
	BucketLocation string

	// Trailer (用于流式签名)
	Trailer http.Header
	AddCRC  bool

	// 特殊处理
	Expect200OKWithError bool
}

// Request 封装的 HTTP 请求
type Request struct {
	ctx      context.Context
	method   string
	metadata RequestMetadata
}

// NewRequest 创建新请求
func NewRequest(ctx context.Context, method string, metadata RequestMetadata) *Request {
	return &Request{
		ctx:      ctx,
		method:   method,
		metadata: metadata,
	}
}

// Context 返回请求上下文
func (r *Request) Context() context.Context {
	return r.ctx
}

// Method 返回 HTTP 方法
func (r *Request) Method() string {
	return r.method
}

// Metadata 返回请求元数据
func (r *Request) Metadata() RequestMetadata {
	return r.metadata
}
