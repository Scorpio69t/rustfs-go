package rustfs

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go/v1/pkg/s3signer"
	"github.com/Scorpio69t/rustfs-go/v1/pkg/s3utils"
)

// MakeBucketOptions - options for MakeBucket
type MakeBucketOptions struct {
	Region string
}

// MakeBucket - create a new bucket
func (c *Client) MakeBucket(ctx context.Context, bucketName string, opts MakeBucketOptions) error {
	// Validate bucket name
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return err
	}

	// Build query values
	queryValues := make(url.Values)

	// Build metadata
	metadata := requestMetadata{
		bucketName:   bucketName,
		queryValues:  queryValues,
		customHeader: make(http.Header),
	}

	// Set region if provided
	if opts.Region != "" {
		metadata.customHeader.Set("x-amz-bucket-region", opts.Region)
	}

	// Execute request
	resp, err := c.executeMethod(ctx, http.MethodPut, metadata)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	return nil
}

// RemoveBucketOptions - options for RemoveBucket
type RemoveBucketOptions struct{}

// RemoveBucket - remove a bucket
func (c *Client) RemoveBucket(ctx context.Context, bucketName string, opts RemoveBucketOptions) error {
	// Validate bucket name
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return err
	}

	// Build metadata
	metadata := requestMetadata{
		bucketName:   bucketName,
		queryValues:  make(url.Values),
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

// BucketInfo - bucket information
type BucketInfo struct {
	Name         string
	CreationDate time.Time
}

// ListBucketsResult - list buckets result
type ListBucketsResult struct {
	XMLName xml.Name     `xml:"ListAllMyBucketsResult"`
	Buckets []BucketInfo `xml:"Buckets>Bucket"`
	Owner   Owner        `xml:"Owner"`
}

// Owner - bucket owner
type Owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

// ListBuckets - list all buckets
func (c *Client) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
	// Build metadata
	metadata := requestMetadata{
		queryValues:  make(url.Values),
		customHeader: make(http.Header),
	}

	// Execute request
	resp, err := c.executeMethod(ctx, http.MethodGet, metadata)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)

	// Parse response
	var result ListBucketsResult
	if err := parseResponse(resp.Body, &result); err != nil {
		return nil, err
	}

	return result.Buckets, nil
}

// BucketExists - check if bucket exists
func (c *Client) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	// Validate bucket name
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return false, err
	}

	// Build metadata
	metadata := requestMetadata{
		bucketName:   bucketName,
		queryValues:  make(url.Values),
		customHeader: make(http.Header),
	}

	// Execute HEAD request
	req, err := c.buildRequest(ctx, http.MethodHead, metadata)
	if err != nil {
		return false, err
	}

	// Sign request
	if c.accessKey != "" && c.secretKey != "" {
		err = s3signer.SignV4(req, c.accessKey, c.secretKey, c.region, "s3", time.Now())
		if err != nil {
			return false, err
		}
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return false, err
	}
	defer closeResponse(resp)

	return resp.StatusCode == http.StatusOK, nil
}

// ListObjectsOptions - options for ListObjects
type ListObjectsOptions struct {
	Prefix    string
	Marker    string
	Delimiter string
	MaxKeys   int
}

// ObjectInfo - object information
type ObjectInfo struct {
	Key          string
	LastModified time.Time
	Size         int64
	ETag         string
	ContentType  string
	Owner        Owner
	StorageClass string
}

// ListObjectsResult - list objects result
type ListObjectsResult struct {
	XMLName        xml.Name     `xml:"ListBucketResult"`
	Name           string       `xml:"Name"`
	Prefix         string       `xml:"Prefix"`
	Marker         string       `xml:"Marker"`
	MaxKeys        int          `xml:"MaxKeys"`
	Delimiter      string       `xml:"Delimiter"`
	IsTruncated    bool         `xml:"IsTruncated"`
	NextMarker     string       `xml:"NextMarker"`
	Contents       []ObjectInfo `xml:"Contents"`
	CommonPrefixes []string     `xml:"CommonPrefixes>Prefix"`
}

// ListObjects - list objects in a bucket
func (c *Client) ListObjects(ctx context.Context, bucketName string, opts ListObjectsOptions) <-chan ObjectInfo {
	objectInfoCh := make(chan ObjectInfo, 1)

	go func() {
		defer close(objectInfoCh)

		queryValues := make(url.Values)
		if opts.Prefix != "" {
			queryValues.Set("prefix", opts.Prefix)
		}
		if opts.Marker != "" {
			queryValues.Set("marker", opts.Marker)
		}
		if opts.Delimiter != "" {
			queryValues.Set("delimiter", opts.Delimiter)
		}
		if opts.MaxKeys > 0 {
			queryValues.Set("max-keys", fmt.Sprintf("%d", opts.MaxKeys))
		}

		metadata := requestMetadata{
			bucketName:   bucketName,
			queryValues:  queryValues,
			customHeader: make(http.Header),
		}

		resp, err := c.executeMethod(ctx, http.MethodGet, metadata)
		if err != nil {
			return
		}
		defer closeResponse(resp)

		var result ListObjectsResult
		if err := parseResponse(resp.Body, &result); err != nil {
			return
		}

		for _, obj := range result.Contents {
			select {
			case objectInfoCh <- obj:
			case <-ctx.Done():
				return
			}
		}
	}()

	return objectInfoCh
}
