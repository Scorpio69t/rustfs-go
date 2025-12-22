// Package bucket bucket/exists.go
package bucket

import (
	"context"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/errors"
	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// Exists checks if a bucket exists.
func (s *bucketService) Exists(ctx context.Context, bucketName string) (bool, error) {
	// validate bucket name
	if err := validateBucketName(bucketName); err != nil {
		return false, err
	}

	// prepare request metadata
	meta := core.RequestMetadata{
		BucketName: bucketName,
	}

	// create request
	req := core.NewRequest(ctx, http.MethodHead, meta)

	// execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		// check if error is NoSuchBucket
		if apiErr, ok := err.(*errors.APIError); ok {
			if apiErr.Code() == errors.ErrCodeNoSuchBucket {
				return false, nil
			}
		}
		return false, err
	}
	defer closeResponse(resp)

	// check response status code
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		err := parseErrorResponse(resp, bucketName, "")
		// check if error is NoSuchBucket
		if apiErr, ok := err.(*errors.APIError); ok {
			if apiErr.Code() == errors.ErrCodeNoSuchBucket {
				return false, nil
			}
		}
		return false, err
	}

	return true, nil
}
