// Package bucket bucket/replication.go
package bucket

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/pkg/replication"
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

// GetReplicationMetrics retrieves replication metrics for a bucket.
func (s *bucketService) GetReplicationMetrics(ctx context.Context, bucketName string) (replication.Metrics, error) {
	if err := validateBucketName(bucketName); err != nil {
		return replication.Metrics{}, err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
		QueryValues:  url.Values{"replication-metrics": {""}},
	}

	req := core.NewRequest(ctx, http.MethodGet, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return replication.Metrics{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return replication.Metrics{}, parseErrorResponse(resp, bucketName, "")
	}

	var metrics replication.Metrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return replication.Metrics{}, err
	}
	return metrics, nil
}
