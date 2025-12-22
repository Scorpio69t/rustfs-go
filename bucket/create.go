// Package bucket bucket/create.go
package bucket

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/errors"
	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// Create bucket
func (s *bucketService) Create(ctx context.Context, bucketName string, opts ...CreateOption) error {
	// validate name
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	// apply options
	options := applyCreateOptions(opts)

	// if region is not set, use default region
	if options.Region == "" {
		options.Region = "us-east-1"
	}

	// prepare request metadata
	meta := core.RequestMetadata{
		BucketName:     bucketName,
		BucketLocation: options.Region,
		CustomHeader:   make(http.Header),
	}

	// set object locking header
	if options.ObjectLocking {
		meta.CustomHeader.Set("x-amz-bucket-object-lock-enabled", "true")
	}

	// set force create header
	if options.ForceCreate {
		meta.CustomHeader.Set("x-rustfs-force-create", "true")
	}

	// if region is not us-east-1, set location constraint
	if options.Region != "us-east-1" && options.Region != "" {
		config := createBucketConfiguration{
			Location: options.Region,
		}

		configBytes, err := xml.Marshal(config)
		if err != nil {
			return err
		}

		meta.ContentBody = bytes.NewReader(configBytes)
		meta.ContentLength = int64(len(configBytes))
		meta.ContentSHA256Hex = sumSHA256Hex(configBytes)
	}

	// prepare request
	req := core.NewRequest(ctx, http.MethodPut, meta)

	// execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	if resp == nil {
		return errors.ErrNilResponse
	}

	// check response status code
	if resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp, bucketName, "")
	}

	// cache location
	if s.locationCache != nil {
		s.locationCache.Set(bucketName, options.Region)
	}

	return nil
}

// createBucketConfiguration creates XML structure for bucket location constraint
type createBucketConfiguration struct {
	XMLName  xml.Name `xml:"CreateBucketConfiguration"`
	Location string   `xml:"LocationConstraint"`
}
