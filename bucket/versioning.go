// Package bucket bucket/versioning.go
package bucket

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// versioningConfiguration represents the XML payload for bucket versioning.
type versioningConfiguration struct {
	XMLName   xml.Name `xml:"VersioningConfiguration"`
	Status    string   `xml:"Status,omitempty"`
	MFADelete string   `xml:"MFADelete,omitempty"`
}

// SetVersioning sets bucket versioning configuration.
func (s *bucketService) SetVersioning(ctx context.Context, bucketName string, cfg types.VersioningConfig) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if cfg.Status != "Enabled" && cfg.Status != "Suspended" {
		return ErrInvalidVersioningStatus
	}

	bodyStruct := versioningConfiguration{
		Status:    cfg.Status,
		MFADelete: cfg.MFADelete,
	}
	body, err := xml.Marshal(bodyStruct)
	if err != nil {
		return err
	}

	meta := core.RequestMetadata{
		BucketName:    bucketName,
		CustomHeader:  make(http.Header),
		QueryValues:   url.Values{"versioning": {""}},
		ContentBody:   bytes.NewReader(body),
		ContentLength: int64(len(body)),
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

// GetVersioning retrieves bucket versioning configuration.
func (s *bucketService) GetVersioning(ctx context.Context, bucketName string) (types.VersioningConfig, error) {
	if err := validateBucketName(bucketName); err != nil {
		return types.VersioningConfig{}, err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
		QueryValues:  url.Values{"versioning": {""}},
	}

	req := core.NewRequest(ctx, http.MethodGet, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return types.VersioningConfig{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return types.VersioningConfig{}, parseErrorResponse(resp, bucketName, "")
	}

	var cfg versioningConfiguration
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&cfg); err != nil {
		return types.VersioningConfig{}, err
	}

	return types.VersioningConfig{
		Status:    cfg.Status,
		MFADelete: cfg.MFADelete,
	}, nil
}
