// Package bucket bucket/object_lock.go
package bucket

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/pkg/objectlock"
)

// SetObjectLockConfig sets the object lock configuration for a bucket.
func (s *bucketService) SetObjectLockConfig(ctx context.Context, bucketName string, config objectlock.Config) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if err := config.Normalize(); err != nil {
		return err
	}

	xmlData, err := xml.Marshal(config)
	if err != nil {
		return err
	}

	meta := core.RequestMetadata{
		BucketName:    bucketName,
		QueryValues:   url.Values{"object-lock": {""}},
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

// GetObjectLockConfig retrieves the object lock configuration for a bucket.
func (s *bucketService) GetObjectLockConfig(ctx context.Context, bucketName string) (objectlock.Config, error) {
	if err := validateBucketName(bucketName); err != nil {
		return objectlock.Config{}, err
	}

	meta := core.RequestMetadata{
		BucketName:  bucketName,
		QueryValues: url.Values{"object-lock": {""}},
	}

	req := core.NewRequest(ctx, http.MethodGet, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return objectlock.Config{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return objectlock.Config{}, objectlock.ErrNoObjectLockConfig
		}
		return objectlock.Config{}, parseErrorResponse(resp, bucketName, "")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return objectlock.Config{}, err
	}

	var config objectlock.Config
	if err := xml.Unmarshal(body, &config); err != nil {
		return objectlock.Config{}, err
	}
	return config, nil
}
