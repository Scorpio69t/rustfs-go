// Package bucket bucket/tagging.go
package bucket

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
)

type tagEntry struct {
	Key   string `xml:"Key"`
	Value string `xml:"Value"`
}

type taggingConfig struct {
	XMLName xml.Name   `xml:"Tagging"`
	Tags    []tagEntry `xml:"TagSet>Tag"`
}

// SetTagging sets tags on a bucket.
func (s *bucketService) SetTagging(ctx context.Context, bucketName string, tags map[string]string) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	cfg := taggingConfig{}
	for k, v := range tags {
		cfg.Tags = append(cfg.Tags, tagEntry{Key: k, Value: v})
	}

	body, err := xml.Marshal(cfg)
	if err != nil {
		return err
	}

	meta := core.RequestMetadata{
		BucketName:    bucketName,
		CustomHeader:  make(http.Header),
		QueryValues:   url.Values{"tagging": {""}},
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

// GetTagging retrieves tags for a bucket.
func (s *bucketService) GetTagging(ctx context.Context, bucketName string) (map[string]string, error) {
	if err := validateBucketName(bucketName); err != nil {
		return nil, err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
		QueryValues:  url.Values{"tagging": {""}},
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

	var cfg taggingConfig
	parser := core.NewResponseParser()
	if err := parser.ParseXML(resp, &cfg); err != nil {
		return nil, err
	}

	result := make(map[string]string, len(cfg.Tags))
	for _, t := range cfg.Tags {
		result[t.Key] = t.Value
	}
	return result, nil
}

// DeleteTagging removes tags from a bucket.
func (s *bucketService) DeleteTagging(ctx context.Context, bucketName string) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
		QueryValues:  url.Values{"tagging": {""}},
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
