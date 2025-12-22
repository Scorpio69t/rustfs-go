// Package object object/delete.go
package object

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// Delete deletes an object (implementation)
func (s *objectService) Delete(ctx context.Context, bucketName, objectName string, opts ...DeleteOption) error {
	// Validate parameters
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if err := validateObjectName(objectName); err != nil {
		return err
	}

	// Apply options
	options := applyDeleteOptions(opts)

	// Build request metadata
	meta := core.RequestMetadata{
		BucketName:   bucketName,
		ObjectName:   objectName,
		CustomHeader: make(http.Header),
	}

	// Add version ID query parameter
	if options.VersionID != "" {
		meta.QueryValues = url.Values{}
		meta.QueryValues.Set("versionId", options.VersionID)
	}

	// Set force delete header (if supported)
	if options.ForceDelete {
		meta.CustomHeader.Set("x-rustfs-force-delete", "true")
	}

	// Merge custom headers
	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	// Create DELETE request
	req := core.NewRequest(ctx, http.MethodDelete, meta)

	// Execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	// Check response (204 No Content or 200 OK)
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp, bucketName, objectName)
	}

	return nil
}
