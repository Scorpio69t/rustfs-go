// Package object object/multipart.go
package object

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// initiateMultipartUploadResult represents the InitiateMultipartUpload response
type initiateMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	UploadID string   `xml:"UploadId"`
}

// completeMultipartUploadResult represents the CompleteMultipartUpload response
type completeMultipartUploadResult struct {
	XMLName  xml.Name  `xml:"CompleteMultipartUploadResult"`
	Location string    `xml:"Location"`
	Bucket   string    `xml:"Bucket"`
	Key      string    `xml:"Key"`
	ETag     string    `xml:"ETag"`
	Modified time.Time `xml:"LastModified,omitempty"`
}

// completeMultipartUpload wraps parts for completion
type completeMultipartUpload struct {
	XMLName xml.Name       `xml:"CompleteMultipartUpload"`
	Parts   []completePart `xml:"Part"`
}

// completePart represents a completed part
type completePart struct {
	PartNumber int    `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

// InitiateMultipartUpload starts a multipart upload
func (s *objectService) InitiateMultipartUpload(ctx context.Context, bucketName, objectName string, opts ...PutOption) (string, error) {
	// Validate inputs
	if err := validateBucketName(bucketName); err != nil {
		return "", err
	}
	if err := validateObjectName(objectName); err != nil {
		return "", err
	}

	// Apply options
	options := applyPutOptions(opts)
	if !options.contentTypeSet && options.ContentType == "" {
		options.ContentType = "application/octet-stream"
	}
	if !options.contentTypeSet && options.ContentType == "" {
		options.ContentType = "application/octet-stream"
	}

	// Build request metadata
	meta := core.RequestMetadata{
		BucketName:   bucketName,
		ObjectName:   objectName,
		QueryValues:  url.Values{},
		CustomHeader: make(http.Header),
	}

	// Add uploads query parameter
	meta.QueryValues.Set("uploads", "")

	// Set Content-Type
	if options.ContentType != "" {
		meta.CustomHeader.Set("Content-Type", options.ContentType)
	}

	// Set Content-Encoding
	if options.ContentEncoding != "" {
		meta.CustomHeader.Set("Content-Encoding", options.ContentEncoding)
	}

	// Set Content-Disposition
	if options.ContentDisposition != "" {
		meta.CustomHeader.Set("Content-Disposition", options.ContentDisposition)
	}

	// Set storage class
	if options.StorageClass != "" {
		meta.CustomHeader.Set("x-amz-storage-class", options.StorageClass)
	}

	// Set user metadata
	if options.UserMetadata != nil {
		for k, v := range options.UserMetadata {
			meta.CustomHeader.Set("x-amz-meta-"+k, v)
		}
	}

	// Set user tags
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

	// Merge custom headers
	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	// Set SSE-C headers if provided (must match initiation)
	applySSECustomerHeaders(&meta, options.SSECustomerAlgorithm, options.SSECustomerKey, options.SSECustomerKeyMD5)

	// Set SSE-C headers if provided
	applySSECustomerHeaders(&meta, options.SSECustomerAlgorithm, options.SSECustomerKey, options.SSECustomerKeyMD5)

	// Create POST request
	req := core.NewRequest(ctx, http.MethodPost, meta)

	// Execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return "", err
	}
	defer closeResponse(resp)

	// Check response
	if resp.StatusCode != http.StatusOK {
		return "", parseErrorResponse(resp, bucketName, objectName)
	}

	// Parse response
	var result initiateMultipartUploadResult
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode initiate multipart upload response: %w", err)
	}

	return result.UploadID, nil
}

// UploadPart uploads a single part
func (s *objectService) UploadPart(ctx context.Context, bucketName, objectName, uploadID string, partNumber int, reader io.Reader, partSize int64, opts ...PutOption) (types.ObjectPart, error) {
	// Validate inputs
	if err := validateBucketName(bucketName); err != nil {
		return types.ObjectPart{}, err
	}
	if err := validateObjectName(objectName); err != nil {
		return types.ObjectPart{}, err
	}
	if uploadID == "" {
		return types.ObjectPart{}, fmt.Errorf("upload ID cannot be empty")
	}
	if partNumber <= 0 {
		return types.ObjectPart{}, fmt.Errorf("part number must be greater than 0")
	}
	if reader == nil {
		return types.ObjectPart{}, fmt.Errorf("reader cannot be nil")
	}

	// Apply options
	options := applyPutOptions(opts)

	// Build request metadata
	meta := core.RequestMetadata{
		BucketName:    bucketName,
		ObjectName:    objectName,
		ContentBody:   reader,
		ContentLength: partSize,
		QueryValues:   url.Values{},
		CustomHeader:  make(http.Header),
	}

	// Set query parameters
	meta.QueryValues.Set("uploadId", uploadID)
	meta.QueryValues.Set("partNumber", strconv.Itoa(partNumber))

	// Merge custom headers
	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	// Set SSE-C headers if provided (required for SSE-C multipart uploads)
	applySSECustomerHeaders(&meta, options.SSECustomerAlgorithm, options.SSECustomerKey, options.SSECustomerKeyMD5)

	// Create PUT request
	req := core.NewRequest(ctx, http.MethodPut, meta)

	// Execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return types.ObjectPart{}, err
	}
	defer closeResponse(resp)

	// Check response
	if resp.StatusCode != http.StatusOK {
		return types.ObjectPart{}, parseErrorResponse(resp, bucketName, objectName)
	}

	// Build part info
	part := types.ObjectPart{
		PartNumber: partNumber,
		ETag:       trimETag(resp.Header.Get("ETag")),
		Size:       partSize,
	}

	// Parse checksum headers
	if checksumCRC32 := resp.Header.Get("x-amz-checksum-crc32"); checksumCRC32 != "" {
		part.ChecksumCRC32 = checksumCRC32
	}
	if checksumCRC32C := resp.Header.Get("x-amz-checksum-crc32c"); checksumCRC32C != "" {
		part.ChecksumCRC32C = checksumCRC32C
	}
	if checksumSHA1 := resp.Header.Get("x-amz-checksum-sha1"); checksumSHA1 != "" {
		part.ChecksumSHA1 = checksumSHA1
	}
	if checksumSHA256 := resp.Header.Get("x-amz-checksum-sha256"); checksumSHA256 != "" {
		part.ChecksumSHA256 = checksumSHA256
	}

	return part, nil
}

// CompleteMultipartUpload finalizes a multipart upload
func (s *objectService) CompleteMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string, parts []types.ObjectPart, opts ...PutOption) (types.UploadInfo, error) {
	// Validate inputs
	if err := validateBucketName(bucketName); err != nil {
		return types.UploadInfo{}, err
	}
	if err := validateObjectName(objectName); err != nil {
		return types.UploadInfo{}, err
	}
	if uploadID == "" {
		return types.UploadInfo{}, fmt.Errorf("upload ID cannot be empty")
	}
	if len(parts) == 0 {
		return types.UploadInfo{}, fmt.Errorf("parts cannot be empty")
	}

	// Apply options
	options := applyPutOptions(opts)

	// Build completion payload
	completeParts := make([]completePart, len(parts))
	for i, part := range parts {
		completeParts[i] = completePart{
			PartNumber: part.PartNumber,
			ETag:       part.ETag,
		}
	}

	completeUpload := completeMultipartUpload{
		Parts: completeParts,
	}

	// Encode XML payload
	xmlData, err := xml.Marshal(completeUpload)
	if err != nil {
		return types.UploadInfo{}, fmt.Errorf("failed to marshal complete multipart upload: %w", err)
	}

	// Build request metadata
	meta := core.RequestMetadata{
		BucketName:    bucketName,
		ObjectName:    objectName,
		ContentBody:   nil, // set below
		ContentLength: int64(len(xmlData)),
		QueryValues:   url.Values{},
		CustomHeader:  make(http.Header),
	}

	// Set query parameters
	meta.QueryValues.Set("uploadId", uploadID)

	// Set Content-Type
	meta.CustomHeader.Set("Content-Type", "application/xml")

	// Merge custom headers
	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	// Create POST request with XML payload
	meta.ContentBody = bytes.NewReader(xmlData)

	req := core.NewRequest(ctx, http.MethodPost, meta)

	// Execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return types.UploadInfo{}, err
	}
	defer closeResponse(resp)

	// Check response
	if resp.StatusCode != http.StatusOK {
		return types.UploadInfo{}, parseErrorResponse(resp, bucketName, objectName)
	}

	// Parse response
	var result completeMultipartUploadResult
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return types.UploadInfo{}, fmt.Errorf("failed to decode complete multipart upload response: %w", err)
	}

	// Build upload info
	uploadInfo := types.UploadInfo{
		Bucket:       result.Bucket,
		Key:          result.Key,
		ETag:         trimETag(result.ETag),
		VersionID:    resp.Header.Get("x-amz-version-id"),
		LastModified: result.Modified,
	}

	// Sum total size
	var totalSize int64
	for _, part := range parts {
		totalSize += part.Size
	}
	uploadInfo.Size = totalSize

	// Parse checksum headers
	if checksumCRC32 := resp.Header.Get("x-amz-checksum-crc32"); checksumCRC32 != "" {
		uploadInfo.ChecksumCRC32 = checksumCRC32
	}
	if checksumCRC32C := resp.Header.Get("x-amz-checksum-crc32c"); checksumCRC32C != "" {
		uploadInfo.ChecksumCRC32C = checksumCRC32C
	}
	if checksumSHA1 := resp.Header.Get("x-amz-checksum-sha1"); checksumSHA1 != "" {
		uploadInfo.ChecksumSHA1 = checksumSHA1
	}
	if checksumSHA256 := resp.Header.Get("x-amz-checksum-sha256"); checksumSHA256 != "" {
		uploadInfo.ChecksumSHA256 = checksumSHA256
	}

	return uploadInfo, nil
}

// AbortMultipartUpload aborts an in-progress multipart upload
func (s *objectService) AbortMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string) error {
	// Validate inputs
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if err := validateObjectName(objectName); err != nil {
		return err
	}
	if uploadID == "" {
		return fmt.Errorf("upload ID cannot be empty")
	}

	// Build request metadata
	meta := core.RequestMetadata{
		BucketName:  bucketName,
		ObjectName:  objectName,
		QueryValues: url.Values{},
	}

	// Set query parameters
	meta.QueryValues.Set("uploadId", uploadID)

	// Create DELETE request
	req := core.NewRequest(ctx, http.MethodDelete, meta)

	// Execute request
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	// Check response (expect 204 No Content)
	if resp.StatusCode != http.StatusNoContent {
		return parseErrorResponse(resp, bucketName, objectName)
	}

	return nil
}
