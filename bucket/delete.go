// Package bucket bucket/delete.go
package bucket

import (
	"context"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/errors"
	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// Delete bucket
func (s *bucketService) Delete(ctx context.Context, bucketName string, opts ...DeleteOption) error {
	// validate bucket name
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	// apply options
	options := applyDeleteOptions(opts)

	// prepare request metadata
	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
	}

	// set force delete header if needed
	if options.ForceDelete {
		meta.CustomHeader.Set("x-rustfs-force-delete", "true")
	}

	// create request
	req := core.NewRequest(ctx, http.MethodDelete, meta)

	// execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	if resp == nil {
		return errors.ErrNilResponse
	}

	// check response status
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp, bucketName, "")
	}

	// delete location cache if exists
	if s.locationCache != nil {
		s.locationCache.Delete(bucketName)
	}

	return nil
}
