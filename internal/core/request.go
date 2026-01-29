// Package core internal/core/request.go
package core

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

// RequestMetadata holds request metadata
type RequestMetadata struct {
	// Bucket and object
	BucketName string
	ObjectName string

	// Query parameters
	QueryValues url.Values

	// Request headers
	CustomHeader http.Header

	// Request body
	ContentBody   io.Reader
	ContentLength int64

	// Content validation
	ContentMD5Base64 string
	ContentSHA256Hex string

	// Signing options
	StreamSHA256 bool
	PresignURL   bool
	Expires      int64

	// Extra headers for presign
	ExtraPresignHeader http.Header

	// Location
	BucketLocation string

	// Trailer (for streaming signature)
	Trailer http.Header
	AddCRC  bool

	// Special handling
	Expect200OKWithError bool

	// Use accelerate endpoint when supported
	UseAccelerate bool
}

// Request encapsulated HTTP request
type Request struct {
	ctx      context.Context
	method   string
	metadata RequestMetadata
}

// NewRequest creates a new Request
func NewRequest(ctx context.Context, method string, metadata RequestMetadata) *Request {
	return &Request{
		ctx:      ctx,
		method:   method,
		metadata: metadata,
	}
}

// Context returns request context
func (r *Request) Context() context.Context {
	return r.ctx
}

// Method returns HTTP method
func (r *Request) Method() string {
	return r.method
}

// Metadata returns request metadata
func (r *Request) Metadata() RequestMetadata {
	return r.metadata
}
