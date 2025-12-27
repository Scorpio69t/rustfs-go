// Package bucket bucket/replication.go
package bucket

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// SetReplication sets the bucket replication configuration (XML).
func (s *bucketService) SetReplication(ctx context.Context, bucketName string, config []byte) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if len(config) == 0 {
		return ErrEmptyBucketConfig
	}

	meta := core.RequestMetadata{
		BucketName:    bucketName,
		CustomHeader:  make(http.Header),
		QueryValues:   url.Values{"replication": {""}},
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

// GetReplication retrieves the bucket replication configuration (XML).
func (s *bucketService) GetReplication(ctx context.Context, bucketName string) ([]byte, error) {
	if err := validateBucketName(bucketName); err != nil {
		return nil, err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
		QueryValues:  url.Values{"replication": {""}},
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

// DeleteReplication removes the bucket replication configuration.
func (s *bucketService) DeleteReplication(ctx context.Context, bucketName string) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
		QueryValues:  url.Values{"replication": {""}},
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
