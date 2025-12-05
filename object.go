package rustfs

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/Scorpio69t/rustfs-go/pkg/s3errors"
	"github.com/Scorpio69t/rustfs-go/pkg/s3signer"
	"github.com/Scorpio69t/rustfs-go/pkg/s3utils"
)

// PutObjectOptions - options for PutObject
type PutObjectOptions struct {
	ContentType     string
	ContentEncoding string
	ContentMD5      string
	UserMetadata    map[string]string
}

// UploadInfo - upload information
type UploadInfo struct {
	Bucket       string
	Key          string
	ETag         string
	Size         int64
	LastModified time.Time
	VersionID    string
}

// PutObject - upload an object
func (c *Client) PutObject(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts PutObjectOptions) (UploadInfo, error) {
	// Validate inputs
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return UploadInfo{}, err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return UploadInfo{}, err
	}

	// Build metadata
	metadata := requestMetadata{
		bucketName:    bucketName,
		objectName:    objectName,
		contentBody:   reader,
		contentLength: objectSize,
		queryValues:   make(url.Values),
		customHeader:  make(http.Header),
	}

	// Set content type
	if opts.ContentType != "" {
		metadata.customHeader.Set("Content-Type", opts.ContentType)
	} else {
		metadata.customHeader.Set("Content-Type", "application/octet-stream")
	}

	// Set content encoding
	if opts.ContentEncoding != "" {
		metadata.customHeader.Set("Content-Encoding", opts.ContentEncoding)
	}

	// Set user metadata
	for k, v := range opts.UserMetadata {
		metadata.customHeader.Set("x-amz-meta-"+k, v)
	}

	// Set content MD5 if provided
	if opts.ContentMD5 != "" {
		metadata.contentMD5Base64 = opts.ContentMD5
	}

	// Execute request
	resp, err := c.executeMethod(ctx, http.MethodPut, metadata)
	if err != nil {
		return UploadInfo{}, err
	}
	defer closeResponse(resp)

	// Parse response
	etag := resp.Header.Get("ETag")
	if len(etag) > 0 && etag[0] == '"' {
		etag = etag[1 : len(etag)-1]
	}

	lastModified, _ := time.Parse(http.TimeFormat, resp.Header.Get("Last-Modified"))

	return UploadInfo{
		Bucket:       bucketName,
		Key:          objectName,
		ETag:         etag,
		Size:         objectSize,
		LastModified: lastModified,
		VersionID:    resp.Header.Get("x-amz-version-id"),
	}, nil
}

// FPutObject - upload an object from file
func (c *Client) FPutObject(ctx context.Context, bucketName, objectName, filePath string, opts PutObjectOptions) (UploadInfo, error) {
	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return UploadInfo{}, err
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return UploadInfo{}, err
	}

	// Calculate MD5 if not provided
	if opts.ContentMD5 == "" {
		hash := md5.New()
		if _, err := io.Copy(hash, file); err != nil {
			return UploadInfo{}, err
		}
		opts.ContentMD5 = base64.StdEncoding.EncodeToString(hash.Sum(nil))
	}

	// Reset file pointer
	if _, err := file.Seek(0, 0); err != nil {
		return UploadInfo{}, err
	}

	// Upload
	return c.PutObject(ctx, bucketName, objectName, file, fileInfo.Size(), opts)
}

// GetObjectOptions - options for GetObject
type GetObjectOptions struct {
	VersionID string
	Range     string
}

// Object - object reader
type Object struct {
	Reader   io.ReadCloser
	Stat     ObjectInfo
	Metadata http.Header
}

// GetObject - download an object
func (c *Client) GetObject(ctx context.Context, bucketName, objectName string, opts GetObjectOptions) (*Object, error) {
	// Validate inputs
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return nil, err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return nil, err
	}

	// Build query values
	queryValues := make(url.Values)
	if opts.VersionID != "" {
		queryValues.Set("versionId", opts.VersionID)
	}

	// Build metadata
	metadata := requestMetadata{
		bucketName:   bucketName,
		objectName:   objectName,
		queryValues:  queryValues,
		customHeader: make(http.Header),
	}

	// Set range header if provided
	if opts.Range != "" {
		metadata.customHeader.Set("Range", opts.Range)
	}

	// Execute request
	resp, err := c.executeMethod(ctx, http.MethodGet, metadata)
	if err != nil {
		return nil, err
	}

	// Parse object info
	lastModified, _ := time.Parse(http.TimeFormat, resp.Header.Get("Last-Modified"))
	size := resp.ContentLength
	if size < 0 {
		size = 0
	}

	etag := resp.Header.Get("ETag")
	if len(etag) > 0 && etag[0] == '"' {
		etag = etag[1 : len(etag)-1]
	}

	objInfo := ObjectInfo{
		Key:          objectName,
		LastModified: lastModified,
		Size:         size,
		ETag:         etag,
		ContentType:  resp.Header.Get("Content-Type"),
		StorageClass: resp.Header.Get("x-amz-storage-class"),
	}

	return &Object{
		Reader:   resp.Body,
		Stat:     objInfo,
		Metadata: resp.Header,
	}, nil
}

// FGetObject - download an object to file
func (c *Client) FGetObject(ctx context.Context, bucketName, objectName, filePath string, opts GetObjectOptions) error {
	// Get object
	obj, err := c.GetObject(ctx, bucketName, objectName, opts)
	if err != nil {
		return err
	}
	defer obj.Reader.Close()

	// Create file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy data
	_, err = io.Copy(file, obj.Reader)
	return err
}

// RemoveObjectOptions - options for RemoveObject
type RemoveObjectOptions struct {
	VersionID string
}

// RemoveObject - delete an object
func (c *Client) RemoveObject(ctx context.Context, bucketName, objectName string, opts RemoveObjectOptions) error {
	// Validate inputs
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return err
	}

	// Build query values
	queryValues := make(url.Values)
	if opts.VersionID != "" {
		queryValues.Set("versionId", opts.VersionID)
	}

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

// RemoveObjectsOptions - options for RemoveObjects
type RemoveObjectsOptions struct {
	VersionID string
}

// RemoveObjectError - error for remove object
type RemoveObjectError struct {
	ObjectName string
	VersionID  string
	Err        error
}

// RemoveObjects - delete multiple objects
func (c *Client) RemoveObjects(ctx context.Context, bucketName string, objectsCh <-chan ObjectInfo, opts RemoveObjectsOptions) <-chan RemoveObjectError {
	errorCh := make(chan RemoveObjectError, 1)

	go func() {
		defer close(errorCh)

		for obj := range objectsCh {
			err := c.RemoveObject(ctx, bucketName, obj.Key, RemoveObjectOptions{
				VersionID: opts.VersionID,
			})
			if err != nil {
				errorCh <- RemoveObjectError{
					ObjectName: obj.Key,
					VersionID:  obj.Key,
					Err:        err,
				}
			}
		}
	}()

	return errorCh
}

// StatObjectOptions - options for StatObject
type StatObjectOptions struct {
	VersionID string
}

// StatObject - get object information
func (c *Client) StatObject(ctx context.Context, bucketName, objectName string, opts StatObjectOptions) (ObjectInfo, error) {
	// Validate inputs
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return ObjectInfo{}, err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return ObjectInfo{}, err
	}

	// Build query values
	queryValues := make(url.Values)
	if opts.VersionID != "" {
		queryValues.Set("versionId", opts.VersionID)
	}

	// Build metadata
	metadata := requestMetadata{
		bucketName:   bucketName,
		objectName:   objectName,
		queryValues:  queryValues,
		customHeader: make(http.Header),
	}

	// Execute HEAD request
	req, err := c.buildRequest(ctx, http.MethodHead, metadata)
	if err != nil {
		return ObjectInfo{}, err
	}

	// Sign request
	if c.accessKey != "" && c.secretKey != "" {
		err = s3signer.SignV4(req, c.accessKey, c.secretKey, c.region, "s3", time.Now())
		if err != nil {
			return ObjectInfo{}, err
		}
	}

	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return ObjectInfo{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ObjectInfo{}, s3errors.ToErrorResponse(resp, bucketName, objectName)
	}

	// Parse response
	lastModified, _ := time.Parse(http.TimeFormat, resp.Header.Get("Last-Modified"))
	size := resp.ContentLength
	if size < 0 {
		size = 0
	}

	etag := resp.Header.Get("ETag")
	if len(etag) > 0 && etag[0] == '"' {
		etag = etag[1 : len(etag)-1]
	}

	return ObjectInfo{
		Key:          objectName,
		LastModified: lastModified,
		Size:         size,
		ETag:         etag,
		ContentType:  resp.Header.Get("Content-Type"),
		StorageClass: resp.Header.Get("x-amz-storage-class"),
	}, nil
}
