// Package bucket bucket/notification.go
package bucket

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// SetNotification sets bucket event notification configuration (XML/JSON).
func (s *bucketService) SetNotification(ctx context.Context, bucketName string, config []byte) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if len(config) == 0 {
		return ErrEmptyBucketConfig
	}

	meta := core.RequestMetadata{
		BucketName:    bucketName,
		CustomHeader:  make(http.Header),
		QueryValues:   url.Values{"notification": {""}},
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

// GetNotification retrieves bucket event notification configuration.
func (s *bucketService) GetNotification(ctx context.Context, bucketName string) ([]byte, error) {
	if err := validateBucketName(bucketName); err != nil {
		return nil, err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
		QueryValues:  url.Values{"notification": {""}},
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

// DeleteNotification removes bucket event notification configuration.
func (s *bucketService) DeleteNotification(ctx context.Context, bucketName string) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
		QueryValues:  url.Values{"notification": {""}},
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
