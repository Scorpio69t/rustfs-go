// Package object object/copy.go
package object

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// copyObjectResult represents copy object result (XML response)
type copyObjectResult struct {
	XMLName      xml.Name  `xml:"CopyObjectResult"`
	ETag         string    `xml:"ETag"`
	LastModified time.Time `xml:"LastModified"`
}

// Copy copies an object (implementation)
func (s *objectService) Copy(ctx context.Context, destBucket, destObject, sourceBucket, sourceObject string, opts ...CopyOption) (types.CopyInfo, error) {
	// Validate parameters
	if err := validateBucketName(destBucket); err != nil {
		return types.CopyInfo{}, err
	}
	if err := validateObjectName(destObject); err != nil {
		return types.CopyInfo{}, err
	}
	if err := validateBucketName(sourceBucket); err != nil {
		return types.CopyInfo{}, err
	}
	if err := validateObjectName(sourceObject); err != nil {
		return types.CopyInfo{}, err
	}

	// Apply options
	options := applyCopyOptions(opts)

	// Build request metadata
	meta := core.RequestMetadata{
		BucketName:   destBucket,
		ObjectName:   destObject,
		CustomHeader: make(http.Header),
		UseAccelerate: options.UseAccelerate,
	}

	// Set copy source
	copySource := fmt.Sprintf("%s/%s", sourceBucket, sourceObject)
	if options.SourceVersionID != "" {
		copySource += "?versionId=" + options.SourceVersionID
	}
	meta.CustomHeader.Set("x-amz-copy-source", copySource)

	// Set metadata directive
	if options.ReplaceMetadata {
		meta.CustomHeader.Set("x-amz-metadata-directive", "REPLACE")

		// Set new metadata
		if options.ContentType != "" {
			meta.CustomHeader.Set("Content-Type", options.ContentType)
		}
		if options.ContentEncoding != "" {
			meta.CustomHeader.Set("Content-Encoding", options.ContentEncoding)
		}
		if options.ContentDisposition != "" {
			meta.CustomHeader.Set("Content-Disposition", options.ContentDisposition)
		}
		if options.CacheControl != "" {
			meta.CustomHeader.Set("Cache-Control", options.CacheControl)
		}
		if !options.Expires.IsZero() {
			meta.CustomHeader.Set("Expires", options.Expires.Format(http.TimeFormat))
		}

		// Set user metadata
		if options.UserMetadata != nil {
			for k, v := range options.UserMetadata {
				meta.CustomHeader.Set("x-amz-meta-"+k, v)
			}
		}
	} else {
		meta.CustomHeader.Set("x-amz-metadata-directive", "COPY")
	}

	// Set tagging directive
	if options.ReplaceTagging {
		meta.CustomHeader.Set("x-amz-tagging-directive", "REPLACE")

		// Set new tags
		if options.UserTags != nil {
			tags := ""
			first := true
			for k, v := range options.UserTags {
				if !first {
					tags += "&"
				}
				tags += k + "=" + v
				first = false
			}
			if tags != "" {
				meta.CustomHeader.Set("x-amz-tagging", tags)
			}
		}
	} else {
		meta.CustomHeader.Set("x-amz-tagging-directive", "COPY")
	}

	// Set conditional match headers (source object)
	if options.MatchETag != "" {
		meta.CustomHeader.Set("x-amz-copy-source-if-match", options.MatchETag)
	}
	if options.NotMatchETag != "" {
		meta.CustomHeader.Set("x-amz-copy-source-if-none-match", options.NotMatchETag)
	}
	if !options.MatchModified.IsZero() {
		meta.CustomHeader.Set("x-amz-copy-source-if-modified-since", options.MatchModified.Format(http.TimeFormat))
	}
	if !options.NotModified.IsZero() {
		meta.CustomHeader.Set("x-amz-copy-source-if-unmodified-since", options.NotModified.Format(http.TimeFormat))
	}

	// Set storage class
	if options.StorageClass != "" {
		meta.CustomHeader.Set("x-amz-storage-class", options.StorageClass)
	}

	// Merge custom headers
	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	// Create PUT request (copy uses PUT method)
	req := core.NewRequest(ctx, http.MethodPut, meta)

	// Execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return types.CopyInfo{}, err
	}
	defer closeResponse(resp)

	// Check response
	if resp.StatusCode != http.StatusOK {
		return types.CopyInfo{}, parseErrorResponse(resp, destBucket, destObject)
	}

	// Parse copy result
	var cpResult copyObjectResult
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&cpResult); err != nil {
		return types.CopyInfo{}, fmt.Errorf("failed to decode copy object response: %w", err)
	}

	// Build copy info
	copyInfo := types.CopyInfo{
		Bucket:          destBucket,
		Key:             destObject,
		ETag:            trimETag(cpResult.ETag),
		LastModified:    cpResult.LastModified,
		VersionID:       resp.Header.Get("x-amz-version-id"),
		SourceVersionID: options.SourceVersionID,
	}

	// Parse checksums
	if checksumCRC32 := resp.Header.Get("x-amz-checksum-crc32"); checksumCRC32 != "" {
		copyInfo.ChecksumCRC32 = checksumCRC32
	}
	if checksumCRC32C := resp.Header.Get("x-amz-checksum-crc32c"); checksumCRC32C != "" {
		copyInfo.ChecksumCRC32C = checksumCRC32C
	}
	if checksumSHA1 := resp.Header.Get("x-amz-checksum-sha1"); checksumSHA1 != "" {
		copyInfo.ChecksumSHA1 = checksumSHA1
	}
	if checksumSHA256 := resp.Header.Get("x-amz-checksum-sha256"); checksumSHA256 != "" {
		copyInfo.ChecksumSHA256 = checksumSHA256
	}

	return copyInfo, nil
}
