// Package object object/presign.go
package object

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// applyPresignOptions applies presign options
func applyPresignOptions(opts []PresignOption) PresignOptions {
	options := PresignOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// PresignGet generates a presigned GET URL for the given object.
// reqParams may include response header overrides such as response-content-type.
func (s *objectService) PresignGet(ctx context.Context, bucketName, objectName string, expires time.Duration, reqParams url.Values, opts ...PresignOption) (*url.URL, http.Header, error) {
	return s.presign(ctx, http.MethodGet, bucketName, objectName, expires, reqParams, opts...)
}

// PresignHead generates a presigned HEAD URL for the given object.
// reqParams may include response header overrides such as response-content-type.
func (s *objectService) PresignHead(ctx context.Context, bucketName, objectName string, expires time.Duration, reqParams url.Values, opts ...PresignOption) (*url.URL, http.Header, error) {
	return s.presign(ctx, http.MethodHead, bucketName, objectName, expires, reqParams, opts...)
}

// PresignPut generates a presigned PUT URL for uploading an object.
// reqParams can include request constraints like content-type.
func (s *objectService) PresignPut(ctx context.Context, bucketName, objectName string, expires time.Duration, reqParams url.Values, opts ...PresignOption) (*url.URL, http.Header, error) {
	return s.presign(ctx, http.MethodPut, bucketName, objectName, expires, reqParams, opts...)
}

// presign is a shared helper that signs a request without sending it.
func (s *objectService) presign(ctx context.Context, method, bucketName, objectName string, expires time.Duration, reqParams url.Values, opts ...PresignOption) (*url.URL, http.Header, error) {
	if err := validateBucketName(bucketName); err != nil {
		return nil, nil, err
	}
	if err := validateObjectName(objectName); err != nil {
		return nil, nil, err
	}
	if expires <= 0 {
		return nil, nil, errors.New("expires must be greater than 0")
	}

	options := applyPresignOptions(opts)

	// Prepare query values (copy to avoid mutation)
	query := url.Values{}
	for k, v := range reqParams {
		for _, vv := range v {
			query.Add(k, vv)
		}
	}
	for k, v := range options.QueryValues {
		for _, vv := range v {
			query.Add(k, vv)
		}
	}

	// Merge signed headers
	headers := make(http.Header)
	for k, v := range options.Headers {
		headers[k] = append([]string{}, v...)
	}

	meta := core.RequestMetadata{
		BucketName:         bucketName,
		ObjectName:         objectName,
		QueryValues:        query,
		CustomHeader:       headers,
		ExtraPresignHeader: headers,
		PresignURL:         true,
		Expires:            int64(expires.Seconds()),
	}

	req := core.NewRequest(ctx, method, meta)
	signedURL, signedHeaders, err := s.executor.Presign(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return signedURL, signedHeaders, nil
}
