// Package bucket bucket/acl.go
package bucket

import (
	"bytes"
	"context"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/pkg/acl"
)

// SetACL sets the ACL for a bucket.
func (s *bucketService) SetACL(ctx context.Context, bucketName string, policy acl.ACL) error {
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if err := policy.Normalize(); err != nil {
		return err
	}

	meta := core.RequestMetadata{
		BucketName:   bucketName,
		QueryValues:  url.Values{"acl": {""}},
		CustomHeader: make(http.Header),
	}

	if policy.Canned != "" {
		meta.CustomHeader.Set("x-amz-acl", string(policy.Canned))
	} else {
		xmlData, err := policy.ToXML()
		if err != nil {
			return err
		}
		meta.ContentBody = bytes.NewReader(xmlData)
		meta.ContentLength = int64(len(xmlData))
		meta.CustomHeader.Set("Content-Type", "application/xml")
	}

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

// GetACL retrieves the ACL for a bucket.
func (s *bucketService) GetACL(ctx context.Context, bucketName string) (acl.ACL, error) {
	if err := validateBucketName(bucketName); err != nil {
		return acl.ACL{}, err
	}

	meta := core.RequestMetadata{
		BucketName:  bucketName,
		QueryValues: url.Values{"acl": {""}},
	}

	req := core.NewRequest(ctx, http.MethodGet, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return acl.ACL{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return acl.ACL{}, parseErrorResponse(resp, bucketName, "")
	}

	policy, err := acl.ParseACL(resp.Body)
	if err != nil {
		return acl.ACL{}, err
	}
	return policy, nil
}
