// Package core internal/core/response.go
package core

import (
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Scorpio69t/rustfs-go/errors"
	"github.com/Scorpio69t/rustfs-go/types"
)

// ResponseParser parses HTTP responses
type ResponseParser struct{}

// NewResponseParser creates a ResponseParser
func NewResponseParser() *ResponseParser {
	return &ResponseParser{}
}

// ParseXML decodes XML response
func (p *ResponseParser) ParseXML(resp *http.Response, v interface{}) error {
	if resp.Body == nil {
		return errors.NewAPIError(errors.ErrCodeInternalError, "empty response body", resp.StatusCode)
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			_ = err
		}
	}(resp.Body)

	return xml.NewDecoder(resp.Body).Decode(v)
}

// ParseObjectInfo parses object info from response headers
func (p *ResponseParser) ParseObjectInfo(resp *http.Response, bucketName, objectName string) (types.ObjectInfo, error) {
	header := resp.Header

	info := types.ObjectInfo{
		Key:         objectName,
		ContentType: header.Get("Content-Type"),
		ETag:        trimETag(header.Get("ETag")),
	}

	// Parse Content-Length
	if cl := header.Get("Content-Length"); cl != "" {
		if size, err := strconv.ParseInt(cl, 10, 64); err == nil {
			info.Size = size
		}
	}

	// Parse Last-Modified
	if lm := header.Get("Last-Modified"); lm != "" {
		if t, err := time.Parse(http.TimeFormat, lm); err == nil {
			info.LastModified = t
		}
	}

	// Parse version info
	info.VersionID = header.Get("x-amz-version-id")
	info.IsDeleteMarker = header.Get("x-amz-delete-marker") == "true"

	// Parse storage class
	info.StorageClass = header.Get("x-amz-storage-class")

	// Parse replication status
	info.ReplicationStatus = header.Get("x-amz-replication-status")

	// Parse user metadata
	info.UserMetadata = make(types.StringMap)
	for k, v := range header {
		if len(k) > len("X-Amz-Meta-") && k[:len("X-Amz-Meta-")] == "X-Amz-Meta-" {
			info.UserMetadata[k[len("X-Amz-Meta-"):]] = v[0]
		}
	}

	// Parse tag count
	if tc := header.Get("x-amz-tagging-count"); tc != "" {
		if count, err := strconv.Atoi(tc); err == nil {
			info.UserTagCount = count
		}
	}

	// Parse checksums
	info.ChecksumCRC32 = header.Get("x-amz-checksum-crc32")
	info.ChecksumCRC32C = header.Get("x-amz-checksum-crc32c")
	info.ChecksumSHA1 = header.Get("x-amz-checksum-sha1")
	info.ChecksumSHA256 = header.Get("x-amz-checksum-sha256")
	info.ChecksumCRC64NVME = header.Get("x-amz-checksum-crc64nvme")

	// Parse restore status
	if restoreHeader := header.Get("x-amz-restore"); restoreHeader != "" {
		info.Restore = parseRestoreHeader(restoreHeader)
	}

	return info, nil
}

// ParseUploadInfo parses upload info from response
func (p *ResponseParser) ParseUploadInfo(resp *http.Response, bucketName, objectName string) (types.UploadInfo, error) {
	header := resp.Header

	info := types.UploadInfo{
		Bucket:    bucketName,
		Key:       objectName,
		ETag:      trimETag(header.Get("ETag")),
		VersionID: header.Get("x-amz-version-id"),
	}

	// Parse checksums
	info.ChecksumCRC32 = header.Get("x-amz-checksum-crc32")
	info.ChecksumCRC32C = header.Get("x-amz-checksum-crc32c")
	info.ChecksumSHA1 = header.Get("x-amz-checksum-sha1")
	info.ChecksumSHA256 = header.Get("x-amz-checksum-sha256")
	info.ChecksumCRC64NVME = header.Get("x-amz-checksum-crc64nvme")

	return info, nil
}

// ParseError parses error response
func (p *ResponseParser) ParseError(resp *http.Response, bucketName, objectName string) error {
	return errors.ParseErrorResponse(resp, bucketName, objectName)
}

// trimETag removes quotes around ETag
func trimETag(etag string) string {
	if len(etag) > 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
		return etag[1 : len(etag)-1]
	}
	return etag
}

func parseRestoreHeader(value string) *types.RestoreInfo {
	if value == "" {
		return nil
	}

	info := &types.RestoreInfo{}
	if idx := strings.Index(value, "ongoing-request=\""); idx >= 0 {
		start := idx + len("ongoing-request=\"")
		if end := strings.Index(value[start:], "\""); end >= 0 {
			info.OngoingRestore = value[start:start+end] == "true"
		}
	}
	if idx := strings.Index(value, "expiry-date=\""); idx >= 0 {
		start := idx + len("expiry-date=\"")
		if end := strings.Index(value[start:], "\""); end >= 0 {
			if parsed, err := time.Parse(http.TimeFormat, value[start:start+end]); err == nil {
				info.ExpiryTime = parsed
			}
		}
	}

	return info
}
