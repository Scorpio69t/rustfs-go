// Package core internal/core/response.go
package core

import (
	"encoding/xml"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/Scorpio69t/rustfs-go/errors"
	"github.com/Scorpio69t/rustfs-go/types"
)

// ResponseParser 响应解析器
type ResponseParser struct{}

// NewResponseParser 创建响应解析器
func NewResponseParser() *ResponseParser {
	return &ResponseParser{}
}

// ParseXML 解析 XML 响应
func (p *ResponseParser) ParseXML(resp *http.Response, v interface{}) error {
	if resp.Body == nil {
		return errors.NewAPIError(errors.ErrCodeInternalError, "empty response body", resp.StatusCode)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			// 记录日志或处理关闭错误
		}
	}(resp.Body)

	return xml.NewDecoder(resp.Body).Decode(v)
}

// ParseObjectInfo 从响应头解析对象信息
func (p *ResponseParser) ParseObjectInfo(resp *http.Response, bucketName, objectName string) (types.ObjectInfo, error) {
	header := resp.Header

	info := types.ObjectInfo{
		Key:         objectName,
		ContentType: header.Get("Content-Type"),
		ETag:        trimETag(header.Get("ETag")),
	}

	// 解析 Content-Length
	if cl := header.Get("Content-Length"); cl != "" {
		if size, err := strconv.ParseInt(cl, 10, 64); err == nil {
			info.Size = size
		}
	}

	// 解析 Last-Modified
	if lm := header.Get("Last-Modified"); lm != "" {
		if t, err := time.Parse(http.TimeFormat, lm); err == nil {
			info.LastModified = t
		}
	}

	// 解析版本信息
	info.VersionID = header.Get("x-amz-version-id")
	info.IsDeleteMarker = header.Get("x-amz-delete-marker") == "true"

	// 解析存储类
	info.StorageClass = header.Get("x-amz-storage-class")

	// 解析复制状态
	info.ReplicationStatus = header.Get("x-amz-replication-status")

	// 解析用户元数据
	info.UserMetadata = make(types.StringMap)
	for k, v := range header {
		if len(k) > len("X-Amz-Meta-") && k[:len("X-Amz-Meta-")] == "X-Amz-Meta-" {
			info.UserMetadata[k[len("X-Amz-Meta-"):]] = v[0]
		}
	}

	// 解析标签数量
	if tc := header.Get("x-amz-tagging-count"); tc != "" {
		if count, err := strconv.Atoi(tc); err == nil {
			info.UserTagCount = count
		}
	}

	// 解析校验和
	info.ChecksumCRC32 = header.Get("x-amz-checksum-crc32")
	info.ChecksumCRC32C = header.Get("x-amz-checksum-crc32c")
	info.ChecksumSHA1 = header.Get("x-amz-checksum-sha1")
	info.ChecksumSHA256 = header.Get("x-amz-checksum-sha256")
	info.ChecksumCRC64NVME = header.Get("x-amz-checksum-crc64nvme")

	return info, nil
}

// ParseUploadInfo 从响应解析上传信息
func (p *ResponseParser) ParseUploadInfo(resp *http.Response, bucketName, objectName string) (types.UploadInfo, error) {
	header := resp.Header

	info := types.UploadInfo{
		Bucket:    bucketName,
		Key:       objectName,
		ETag:      trimETag(header.Get("ETag")),
		VersionID: header.Get("x-amz-version-id"),
	}

	// 解析校验和
	info.ChecksumCRC32 = header.Get("x-amz-checksum-crc32")
	info.ChecksumCRC32C = header.Get("x-amz-checksum-crc32c")
	info.ChecksumSHA1 = header.Get("x-amz-checksum-sha1")
	info.ChecksumSHA256 = header.Get("x-amz-checksum-sha256")
	info.ChecksumCRC64NVME = header.Get("x-amz-checksum-crc64nvme")

	return info, nil
}

// ParseError 解析错误响应
func (p *ResponseParser) ParseError(resp *http.Response, bucketName, objectName string) error {
	return errors.ParseErrorResponse(resp, bucketName, objectName)
}

// trimETag 去除 ETag 的引号
func trimETag(etag string) string {
	if len(etag) > 2 && etag[0] == '"' && etag[len(etag)-1] == '"' {
		return etag[1 : len(etag)-1]
	}
	return etag
}
