// Package bucket bucket/logging.go
package bucket

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// SetLogging sets bucket access logging configuration (XML).
func (s *bucketService) SetLogging(ctx context.Context, bucketName string, config []byte) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if len(config) == 0 {
		return ErrEmptyBucketConfig
	}

	meta := core.RequestMetadata{
		BucketName:    bucketName,
		CustomHeader:  make(http.Header),
		QueryValues:   url.Values{"logging": {""}},
		ContentBody:   bytes.NewReader(config),
		ContentLength: int64(len(config)),
	}
	meta.CustomHeader.Set("Content-Type", "application/xml")

	req := core.NewRequest(ctx, http.MethodPut, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return parseErrorResponse(resp, bucketName, "")
	}
	return nil
}

// GetLogging retrieves bucket access logging configuration (XML).
func (s *bucketService) GetLogging(ctx context.Context, bucketName string) ([]byte, error) {
	if err := validateBucketName(bucketName); err != nil {
		return nil, err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
		QueryValues:  url.Values{"logging": {""}},
	}

	req := core.NewRequest(ctx, http.MethodGet, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, parseErrorResponse(resp, bucketName, "")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// DeleteLogging removes bucket access logging configuration.
func (s *bucketService) DeleteLogging(ctx context.Context, bucketName string) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
		QueryValues:  url.Values{"logging": {""}},
	}

	req := core.NewRequest(ctx, http.MethodDelete, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return parseErrorResponse(resp, bucketName, "")
	}
	return nil
}
