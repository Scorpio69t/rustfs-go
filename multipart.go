package rustfs

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

	"github.com/Scorpio69t/rustfs-go/v1/pkg/s3utils"
)

// InitiateMultipartUploadResult - initiate multipart upload result
type InitiateMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	UploadID string   `xml:"UploadId"`
}

// InitiateMultipartUpload - initiate multipart upload
func (c *Client) InitiateMultipartUpload(ctx context.Context, bucketName, objectName string, opts PutObjectOptions) (string, error) {
	// Validate inputs
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return "", err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return "", err
	}

	// Build query values
	queryValues := make(url.Values)
	queryValues.Set("uploads", "")

	// Build metadata
	metadata := requestMetadata{
		bucketName:   bucketName,
		objectName:   objectName,
		queryValues:  queryValues,
		customHeader: make(http.Header),
	}

	// Set content type
	if opts.ContentType != "" {
		metadata.customHeader.Set("Content-Type", opts.ContentType)
	}

	// Set user metadata
	for k, v := range opts.UserMetadata {
		metadata.customHeader.Set("x-amz-meta-"+k, v)
	}

	// Execute request
	resp, err := c.executeMethod(ctx, http.MethodPost, metadata)
	if err != nil {
		return "", err
	}
	defer closeResponse(resp)

	// Parse response
	var result InitiateMultipartUploadResult
	if err := parseResponse(resp.Body, &result); err != nil {
		return "", err
	}

	return result.UploadID, nil
}

// PutObjectPartOptions - options for UploadPart
type PutObjectPartOptions struct {
	ContentMD5 string
}

// UploadPartInfo - upload part information
type UploadPartInfo struct {
	PartNumber int
	ETag       string
	Size       int64
}

// UploadPart - upload a part
func (c *Client) UploadPart(ctx context.Context, bucketName, objectName, uploadID string, partNumber int, data io.Reader, partSize int64, opts PutObjectPartOptions) (UploadPartInfo, error) {
	// Validate inputs
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return UploadPartInfo{}, err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return UploadPartInfo{}, err
	}
	if err := s3utils.ValidatePartSize(partSize); err != nil {
		return UploadPartInfo{}, err
	}
	if partNumber < 1 || partNumber > 10000 {
		return UploadPartInfo{}, fmt.Errorf("part number must be between 1 and 10000")
	}

	// Build query values
	queryValues := make(url.Values)
	queryValues.Set("partNumber", strconv.Itoa(partNumber))
	queryValues.Set("uploadId", uploadID)

	// Build metadata
	metadata := requestMetadata{
		bucketName:    bucketName,
		objectName:    objectName,
		contentBody:   data,
		contentLength: partSize,
		queryValues:   queryValues,
		customHeader:  make(http.Header),
	}

	// Set content MD5 if provided
	if opts.ContentMD5 != "" {
		metadata.contentMD5Base64 = opts.ContentMD5
	}

	// Execute request
	resp, err := c.executeMethod(ctx, http.MethodPut, metadata)
	if err != nil {
		return UploadPartInfo{}, err
	}
	defer closeResponse(resp)

	// Parse ETag
	etag := resp.Header.Get("ETag")
	if len(etag) > 0 && etag[0] == '"' {
		etag = etag[1 : len(etag)-1]
	}

	return UploadPartInfo{
		PartNumber: partNumber,
		ETag:       etag,
		Size:       partSize,
	}, nil
}

// CompletePart - complete part information
type CompletePart struct {
	PartNumber int    `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

// CompleteMultipartUploadResult - complete multipart upload result
type CompleteMultipartUploadResult struct {
	XMLName      xml.Name `xml:"CompleteMultipartUploadResult"`
	Location     string   `xml:"Location"`
	Bucket       string   `xml:"Bucket"`
	Key          string   `xml:"Key"`
	ETag         string   `xml:"ETag"`
	VersionID    string   `xml:"VersionId"`
	LastModified time.Time
}

// CompleteMultipartUpload - complete multipart upload
func (c *Client) CompleteMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string, parts []CompletePart, opts PutObjectOptions) (UploadInfo, error) {
	// Validate inputs
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return UploadInfo{}, err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return UploadInfo{}, err
	}

	// Build query values
	queryValues := make(url.Values)
	queryValues.Set("uploadId", uploadID)

	// Build complete multipart upload request body
	completeRequest := struct {
		XMLName xml.Name       `xml:"CompleteMultipartUpload"`
		Parts   []CompletePart `xml:"Part"`
	}{
		Parts: parts,
	}

	body, err := xml.Marshal(completeRequest)
	if err != nil {
		return UploadInfo{}, err
	}

	// Build metadata
	metadata := requestMetadata{
		bucketName:    bucketName,
		objectName:    objectName,
		contentBody:   io.NopCloser(bytes.NewReader(body)),
		contentLength: int64(len(body)),
		queryValues:   queryValues,
		customHeader:  make(http.Header),
	}
	metadata.customHeader.Set("Content-Type", "application/xml")

	// Execute request
	resp, err := c.executeMethod(ctx, http.MethodPost, metadata)
	if err != nil {
		return UploadInfo{}, err
	}
	defer closeResponse(resp)

	// Parse response
	var result CompleteMultipartUploadResult
	if err := parseResponse(resp.Body, &result); err != nil {
		return UploadInfo{}, err
	}

	etag := result.ETag
	if len(etag) > 0 && etag[0] == '"' {
		etag = etag[1 : len(etag)-1]
	}

	return UploadInfo{
		Bucket:       bucketName,
		Key:          objectName,
		ETag:         etag,
		LastModified: result.LastModified,
		VersionID:    result.VersionID,
	}, nil
}

// AbortMultipartUploadOptions - options for AbortMultipartUpload
type AbortMultipartUploadOptions struct{}

// AbortMultipartUpload - abort multipart upload
func (c *Client) AbortMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string, opts AbortMultipartUploadOptions) error {
	// Validate inputs
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return err
	}

	// Build query values
	queryValues := make(url.Values)
	queryValues.Set("uploadId", uploadID)

	// Build metadata
	metadata := requestMetadata{
		bucketName:   bucketName,
		objectName:   objectName,
		queryValues:  queryValues,
		customHeader: make(http.Header),
	}

	// Execute request
	resp, err := c.executeMethod(ctx, http.MethodDelete, metadata)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	return nil
}
