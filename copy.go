package rustfs

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go/v1/pkg/s3utils"
)

// CopyObjectOptions - options for CopyObject
type CopyObjectOptions struct {
	Metadata     map[string]string
	ContentType  string
	StorageClass string
}

// CopyObjectInfo - copy object information
type CopyObjectInfo struct {
	ETag         string
	LastModified time.Time
	VersionID    string
}

// CopyObject - copy an object
func (c *Client) CopyObject(ctx context.Context, srcBucket, srcObject, destBucket, destObject string, opts CopyObjectOptions) (CopyObjectInfo, error) {
	// Validate inputs
	if err := s3utils.CheckValidBucketName(srcBucket); err != nil {
		return CopyObjectInfo{}, err
	}
	if err := s3utils.CheckValidBucketName(destBucket); err != nil {
		return CopyObjectInfo{}, err
	}
	if err := s3utils.CheckValidObjectName(srcObject); err != nil {
		return CopyObjectInfo{}, err
	}
	if err := s3utils.CheckValidObjectName(destObject); err != nil {
		return CopyObjectInfo{}, err
	}

	// Build copy source
	copySource := fmt.Sprintf("/%s/%s", srcBucket, srcObject)

	// Build metadata
	metadata := requestMetadata{
		bucketName:   destBucket,
		objectName:   destObject,
		queryValues:  make(url.Values),
		customHeader: make(http.Header),
	}

	// Set copy source header
	metadata.customHeader.Set("x-amz-copy-source", copySource)

	// Set content type
	if opts.ContentType != "" {
		metadata.customHeader.Set("Content-Type", opts.ContentType)
	}

	// Set storage class
	if opts.StorageClass != "" {
		metadata.customHeader.Set("x-amz-storage-class", opts.StorageClass)
	}

	// Set metadata
	for k, v := range opts.Metadata {
		metadata.customHeader.Set("x-amz-meta-"+k, v)
	}

	// Execute request
	resp, err := c.executeMethod(ctx, http.MethodPut, metadata)
	if err != nil {
		return CopyObjectInfo{}, err
	}
	defer closeResponse(resp)

	// Parse response
	var result struct {
		XMLName      xml.Name `xml:"CopyObjectResult"`
		ETag         string   `xml:"ETag"`
		LastModified string   `xml:"LastModified"`
	}
	if err := parseResponse(resp.Body, &result); err != nil {
		return CopyObjectInfo{}, err
	}

	// Parse last modified
	lastModified, _ := time.Parse(time.RFC3339, result.LastModified)

	etag := result.ETag
	if len(etag) > 0 && etag[0] == '"' {
		etag = etag[1 : len(etag)-1]
	}

	return CopyObjectInfo{
		ETag:         etag,
		LastModified: lastModified,
		VersionID:    resp.Header.Get("x-amz-version-id"),
	}, nil
}
