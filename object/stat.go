// Package object object/stat.go
package object

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// Stat gets object information (implementation)
func (s *objectService) Stat(ctx context.Context, bucketName, objectName string, opts ...StatOption) (types.ObjectInfo, error) {
	// Validate parameters
	if err := validateBucketName(bucketName); err != nil {
		return types.ObjectInfo{}, err
	}
	if err := validateObjectName(objectName); err != nil {
		return types.ObjectInfo{}, err
	}

	// Apply options
	options := applyStatOptions(opts)

	// Build request metadata
	meta := core.RequestMetadata{
		BucketName:   bucketName,
		ObjectName:   objectName,
		CustomHeader: options.CustomHeaders,
		UseAccelerate: options.UseAccelerate,
	}

	// Add version ID query parameter
	if options.VersionID != "" {
		meta.QueryValues = url.Values{}
		meta.QueryValues.Set("versionId", options.VersionID)
	}

	// Create HEAD request
	req := core.NewRequest(ctx, http.MethodHead, meta)

	// Execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return types.ObjectInfo{}, err
	}
	defer closeResponse(resp)

	// Check response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return types.ObjectInfo{}, parseErrorResponse(resp, bucketName, objectName)
	}

	// Parse object info
	parser := core.NewResponseParser()
	return parser.ParseObjectInfo(resp, bucketName, objectName)
}
