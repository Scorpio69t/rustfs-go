// Package bucket bucket/cors.go
package bucket

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/pkg/cors"
)

// SetCORS sets the CORS configuration for a bucket.
func (s *bucketService) SetCORS(ctx context.Context, bucketName string, config cors.Config) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	xmlData, err := config.ToXML()
	if err != nil {
		return err
	}

	meta := core.RequestMetadata{
		BucketName:    bucketName,
		QueryValues:   url.Values{"cors": {""}},
		ContentBody:   bytes.NewReader(xmlData),
		ContentLength: int64(len(xmlData)),
		CustomHeader:  make(http.Header),
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

// GetCORS retrieves the CORS configuration for a bucket.
func (s *bucketService) GetCORS(ctx context.Context, bucketName string) (cors.Config, error) {
	if err := validateBucketName(bucketName); err != nil {
		return cors.Config{}, err
	}

	meta := core.RequestMetadata{
		BucketName:  bucketName,
		QueryValues: url.Values{"cors": {""}},
	}

	req := core.NewRequest(ctx, http.MethodGet, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return cors.Config{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return cors.Config{}, cors.ErrNoCORSConfig
		}
		return cors.Config{}, parseErrorResponse(resp, bucketName, "")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return cors.Config{}, err
	}

	cfg, err := cors.ParseBucketCORSConfig(bytes.NewReader(body))
	if err != nil {
		return cors.Config{}, err
	}
	return cfg, nil
}

// DeleteCORS removes the CORS configuration from a bucket.
func (s *bucketService) DeleteCORS(ctx context.Context, bucketName string) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	meta := core.RequestMetadata{
		BucketName:  bucketName,
		QueryValues: url.Values{"cors": {""}},
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
