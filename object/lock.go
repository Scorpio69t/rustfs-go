// Package object object/lock.go
package object

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/http"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/pkg/objectlock"
)

// SetLegalHold sets the legal hold status for an object.
func (s *objectService) SetLegalHold(ctx context.Context, bucketName, objectName string, hold objectlock.LegalHoldStatus, opts ...LegalHoldOption) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if err := validateObjectName(objectName); err != nil {
		return err
	}
	if !hold.IsValid() {
		return objectlock.ErrInvalidLegalHoldStatus
	}

	options := applyLegalHoldOptions(opts)

	cfg := objectlock.LegalHold{Status: hold}
	body, err := xml.Marshal(cfg)
	if err != nil {
		return err
	}

	query := url.Values{"legal-hold": {""}}
	if options.VersionID != "" {
		query.Set("versionId", options.VersionID)
	}

	meta := core.RequestMetadata{
		BucketName:    bucketName,
		ObjectName:    objectName,
		CustomHeader:  make(http.Header),
		QueryValues:   query,
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
		return parseErrorResponse(resp, bucketName, objectName)
	}
	return nil
}

// GetLegalHold retrieves the legal hold status for an object.
func (s *objectService) GetLegalHold(ctx context.Context, bucketName, objectName string, opts ...LegalHoldOption) (objectlock.LegalHoldStatus, error) {
	if err := validateBucketName(bucketName); err != nil {
		return "", err
	}
	if err := validateObjectName(objectName); err != nil {
		return "", err
	}

	options := applyLegalHoldOptions(opts)
	query := url.Values{"legal-hold": {""}}
	if options.VersionID != "" {
		query.Set("versionId", options.VersionID)
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		ObjectName:   objectName,
		CustomHeader: make(http.Header),
		QueryValues:  query,
	}

	req := core.NewRequest(ctx, http.MethodGet, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return "", err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return "", parseErrorResponse(resp, bucketName, objectName)
	}

	var cfg objectlock.LegalHold
	parser := core.NewResponseParser()
	if err := parser.ParseXML(resp, &cfg); err != nil {
		return "", err
	}
	return cfg.Status, nil
}

// SetRetention sets retention mode and retain-until date for an object.
func (s *objectService) SetRetention(ctx context.Context, bucketName, objectName string, mode objectlock.RetentionMode, retainUntil time.Time, opts ...RetentionOption) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if err := validateObjectName(objectName); err != nil {
		return err
	}
	if !mode.IsValid() {
		return objectlock.ErrInvalidRetentionMode
	}
	if retainUntil.IsZero() {
		return objectlock.ErrInvalidRetentionDate
	}

	options := applyRetentionOptions(opts)

	cfg := objectlock.Retention{
		Mode:            mode,
		RetainUntilDate: retainUntil,
	}
	body, err := xml.Marshal(cfg)
	if err != nil {
		return err
	}

	query := url.Values{"retention": {""}}
	if options.VersionID != "" {
		query.Set("versionId", options.VersionID)
	}

	meta := core.RequestMetadata{
		BucketName:    bucketName,
		ObjectName:    objectName,
		CustomHeader:  make(http.Header),
		QueryValues:   query,
		ContentBody:   bytes.NewReader(body),
		ContentLength: int64(len(body)),
	}
	meta.CustomHeader.Set("Content-Type", "application/xml")
	if options.GovernanceBypass {
		meta.CustomHeader.Set("x-amz-bypass-governance-retention", "true")
	}

	req := core.NewRequest(ctx, http.MethodPut, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return parseErrorResponse(resp, bucketName, objectName)
	}
	return nil
}

// GetRetention retrieves retention configuration for an object.
func (s *objectService) GetRetention(ctx context.Context, bucketName, objectName string, opts ...RetentionOption) (objectlock.RetentionMode, time.Time, error) {
	if err := validateBucketName(bucketName); err != nil {
		return "", time.Time{}, err
	}
	if err := validateObjectName(objectName); err != nil {
		return "", time.Time{}, err
	}

	options := applyRetentionOptions(opts)
	query := url.Values{"retention": {""}}
	if options.VersionID != "" {
		query.Set("versionId", options.VersionID)
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		ObjectName:   objectName,
		CustomHeader: make(http.Header),
		QueryValues:  query,
	}

	req := core.NewRequest(ctx, http.MethodGet, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, parseErrorResponse(resp, bucketName, objectName)
	}

	var cfg objectlock.Retention
	parser := core.NewResponseParser()
	if err := parser.ParseXML(resp, &cfg); err != nil {
		return "", time.Time{}, err
	}
	if cfg.Mode != "" && !cfg.Mode.IsValid() {
		return "", time.Time{}, objectlock.ErrInvalidRetentionMode
	}
	return cfg.Mode, cfg.RetainUntilDate, nil
}
