// Package object object/append.go
package object

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// Append appends data to an existing object at the provided offset.
func (s *objectService) Append(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, offset int64, opts ...PutOption) (types.UploadInfo, error) {
	if err := validateBucketName(bucketName); err != nil {
		return types.UploadInfo{}, err
	}
	if err := validateObjectName(objectName); err != nil {
		return types.UploadInfo{}, err
	}
	if reader == nil {
		return types.UploadInfo{}, ErrInvalidObjectName
	}
	if objectSize < 0 {
		return types.UploadInfo{}, fmt.Errorf("object size must be non-negative")
	}
	if offset < 0 {
		info, err := s.Stat(ctx, bucketName, objectName)
		if err != nil {
			return types.UploadInfo{}, err
		}
		offset = info.Size
	}

	options := applyPutOptions(opts)
	if !options.contentTypeSet && options.ContentType == "" {
		options.ContentType = "application/octet-stream"
	}

	meta := core.RequestMetadata{
		BucketName:    bucketName,
		ObjectName:    objectName,
		ContentBody:   reader,
		ContentLength: objectSize,
		CustomHeader:  make(http.Header),
		UseAccelerate: options.UseAccelerate,
	}

	if options.ContentType != "" {
		meta.CustomHeader.Set("Content-Type", options.ContentType)
	}
	if options.ContentEncoding != "" {
		meta.CustomHeader.Set("Content-Encoding", options.ContentEncoding)
	}
	if options.ContentDisposition != "" {
		meta.CustomHeader.Set("Content-Disposition", options.ContentDisposition)
	}
	if options.ContentLanguage != "" {
		meta.CustomHeader.Set("Content-Language", options.ContentLanguage)
	}
	if options.CacheControl != "" {
		meta.CustomHeader.Set("Cache-Control", options.CacheControl)
	}
	if !options.Expires.IsZero() {
		meta.CustomHeader.Set("Expires", options.Expires.Format(http.TimeFormat))
	}
	if options.StorageClass != "" {
		meta.CustomHeader.Set("x-amz-storage-class", options.StorageClass)
	}
	if options.UserMetadata != nil {
		for k, v := range options.UserMetadata {
			meta.CustomHeader.Set("x-amz-meta-"+k, v)
		}
	}
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

	applyChecksumHeaders(&meta, options)

	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	if options.SSE != nil {
		options.SSE.ApplyHeaders(meta.CustomHeader)
	} else {
		applySSECustomerHeaders(&meta, options.SSECustomerAlgorithm, options.SSECustomerKey, options.SSECustomerKeyMD5)
	}

	meta.CustomHeader.Set("x-amz-write-offset-bytes", strconv.FormatInt(offset, 10))

	req := core.NewRequest(ctx, http.MethodPut, meta)
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return types.UploadInfo{}, err
	}
	defer closeResponse(resp)

	if resp.StatusCode != http.StatusOK {
		return types.UploadInfo{}, parseErrorResponse(resp, bucketName, objectName)
	}

	objSizeHeader := resp.Header.Get("x-amz-object-size")
	if objSizeHeader == "" {
		return types.UploadInfo{}, fmt.Errorf("server does not report appended object size")
	}
	finalSize, err := strconv.ParseInt(objSizeHeader, 10, 64)
	if err != nil {
		return types.UploadInfo{}, err
	}
	if finalSize != offset+objectSize {
		return types.UploadInfo{}, fmt.Errorf("server returned incorrect object size")
	}

	parser := core.NewResponseParser()
	uploadInfo, err := parser.ParseUploadInfo(resp, bucketName, objectName)
	if err != nil {
		return types.UploadInfo{}, err
	}
	uploadInfo.Size = finalSize
	uploadInfo.ChecksumMode = resp.Header.Get("x-amz-checksum-mode")

	return uploadInfo, nil
}
