// Package object object/get.go
package object

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// Get downloads an object (implementation)
func (s *objectService) Get(ctx context.Context, bucketName, objectName string, opts ...GetOption) (io.ReadCloser, types.ObjectInfo, error) {
	// Validate parameters
	if err := validateBucketName(bucketName); err != nil {
		return nil, types.ObjectInfo{}, err
	}
	if err := validateObjectName(objectName); err != nil {
		return nil, types.ObjectInfo{}, err
	}

	// Apply options
	options := applyGetOptions(opts)

	// Build request metadata
	meta := core.RequestMetadata{
		BucketName:   bucketName,
		ObjectName:   objectName,
		CustomHeader: make(http.Header),
		UseAccelerate: options.UseAccelerate,
	}

	// Set Range header
	if options.SetRange {
		rangeHeader := "bytes=" + strconv.FormatInt(options.RangeStart, 10) + "-"
		if options.RangeEnd > 0 {
			rangeHeader += strconv.FormatInt(options.RangeEnd, 10)
		}
		meta.CustomHeader.Set("Range", rangeHeader)
	}

	// Set conditional match headers
	if options.MatchETag != "" {
		meta.CustomHeader.Set("If-Match", options.MatchETag)
	}
	if options.NotMatchETag != "" {
		meta.CustomHeader.Set("If-None-Match", options.NotMatchETag)
	}
	if !options.MatchModified.IsZero() {
		meta.CustomHeader.Set("If-Modified-Since", options.MatchModified.Format(http.TimeFormat))
	}
	if !options.NotModified.IsZero() {
		meta.CustomHeader.Set("If-Unmodified-Since", options.NotModified.Format(http.TimeFormat))
	}

	// Add version ID query parameter
	queryValues := url.Values{}
	if options.VersionID != "" {
		queryValues.Set("versionId", options.VersionID)
	}
	for k, v := range options.ResponseHeaderOverrides {
		for _, vv := range v {
			queryValues.Add(k, vv)
		}
	}
	if len(queryValues) > 0 {
		meta.QueryValues = queryValues
	}

	// Set SSE-C headers when downloading encrypted objects
	if options.SSE != nil {
		options.SSE.ApplyHeaders(meta.CustomHeader)
	} else if options.SSECustomerAlgorithm != "" && options.SSECustomerKey != "" {
		// Fallback to legacy SSE-C headers
		meta.CustomHeader.Set("x-amz-server-side-encryption-customer-algorithm", options.SSECustomerAlgorithm)
		meta.CustomHeader.Set("x-amz-server-side-encryption-customer-key", options.SSECustomerKey)
		if options.SSECustomerKeyMD5 != "" {
			meta.CustomHeader.Set("x-amz-server-side-encryption-customer-key-MD5", options.SSECustomerKeyMD5)
		}
	}

	// Merge custom headers
	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	// Create GET request
	req := core.NewRequest(ctx, http.MethodGet, meta)

	// Execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return nil, types.ObjectInfo{}, err
	}

	// Check response (200 OK or 206 Partial Content)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		closeResponse(resp)
		return nil, types.ObjectInfo{}, parseErrorResponse(resp, bucketName, objectName)
	}

	// Parse object info
	parser := core.NewResponseParser()
	objectInfo, err := parser.ParseObjectInfo(resp, bucketName, objectName)
	if err != nil {
		closeResponse(resp)
		return nil, types.ObjectInfo{}, err
	}

	if options.CSE != nil {
		cseMetadata := map[string]string{}
		for key, value := range objectInfo.UserMetadata {
			cseMetadata[strings.ToLower(key)] = value
		}
		decrypted, err := options.CSE.Decrypt(resp.Body, cseMetadata)
		if err != nil {
			closeResponse(resp)
			return nil, types.ObjectInfo{}, err
		}
		closeResponse(resp)
		objectInfo.Size = int64(len(decrypted))
		return io.NopCloser(bytes.NewReader(decrypted)), objectInfo, nil
	}

	// Return response body and object info
	// Note: Caller is responsible for closing Body
	return resp.Body, objectInfo, nil
}
