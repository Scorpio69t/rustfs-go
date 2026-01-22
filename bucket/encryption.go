// Package bucket bucket/encryption.go
package bucket

import (
	"bytes"
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/pkg/sse"
)

// SetEncryption sets default encryption configuration for a bucket
func (s *bucketService) SetEncryption(ctx context.Context, bucketName string, config sse.Configuration) error {
	// Validate bucket name
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	// Marshal configuration to XML
	xmlData, err := xml.Marshal(config)
	if err != nil {
		return err
	}

	// Build request metadata
	meta := core.RequestMetadata{
		BucketName:    bucketName,
		QueryValues:   url.Values{},
		ContentBody:   bytes.NewReader(xmlData),
		ContentLength: int64(len(xmlData)),
		CustomHeader:  make(http.Header),
	}
	meta.QueryValues.Set("encryption", "")
	meta.CustomHeader.Set("Content-Type", "application/xml")

	// Create PUT request
	req := core.NewRequest(ctx, http.MethodPut, meta)

	// Execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	// Check response
	if resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp, bucketName, "")
	}

	return nil
}

// GetEncryption retrieves the default encryption configuration of a bucket
func (s *bucketService) GetEncryption(ctx context.Context, bucketName string) (sse.Configuration, error) {
	// Validate bucket name
	if err := validateBucketName(bucketName); err != nil {
		return sse.Configuration{}, err
	}

	// Build request metadata
	meta := core.RequestMetadata{
		BucketName:  bucketName,
		QueryValues: url.Values{},
	}
	meta.QueryValues.Set("encryption", "")

	// Create GET request
	req := core.NewRequest(ctx, http.MethodGet, meta)

	// Execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return sse.Configuration{}, err
	}
	defer closeResponse(resp)

	// Check response
	if resp.StatusCode != http.StatusOK {
		// Check if encryption not set
		if resp.StatusCode == http.StatusNotFound {
			return sse.Configuration{}, sse.ErrNoEncryptionConfig
		}
		return sse.Configuration{}, parseErrorResponse(resp, bucketName, "")
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return sse.Configuration{}, err
	}

	// Parse XML response
	var config sse.Configuration
	if err := xml.Unmarshal(body, &config); err != nil {
		return sse.Configuration{}, err
	}

	return config, nil
}

// DeleteEncryption removes the default encryption configuration from a bucket
func (s *bucketService) DeleteEncryption(ctx context.Context, bucketName string) error {
	// Validate bucket name
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	// Build request metadata
	meta := core.RequestMetadata{
		BucketName:  bucketName,
		QueryValues: url.Values{},
	}
	meta.QueryValues.Set("encryption", "")

	// Create DELETE request
	req := core.NewRequest(ctx, http.MethodDelete, meta)

	// Execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	// Check response
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp, bucketName, "")
	}

	return nil
}
