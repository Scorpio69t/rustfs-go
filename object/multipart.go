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

// initiateMultipartUploadResult 初始化分片上传响应
type initiateMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	UploadID string   `xml:"UploadId"`
}

// completeMultipartUploadResult 完成分片上传响应
type completeMultipartUploadResult struct {
	XMLName  xml.Name  `xml:"CompleteMultipartUploadResult"`
	Location string    `xml:"Location"`
	Bucket   string    `xml:"Bucket"`
	Key      string    `xml:"Key"`
	ETag     string    `xml:"ETag"`
	Modified time.Time `xml:"LastModified,omitempty"`
}

// completeMultipartUpload 完成分片上传请求
type completeMultipartUpload struct {
	XMLName xml.Name       `xml:"CompleteMultipartUpload"`
	Parts   []completePart `xml:"Part"`
}

// completePart 完成的分片信息
type completePart struct {
	PartNumber int    `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

// InitiateMultipartUpload 初始化分片上传
func (s *objectService) InitiateMultipartUpload(ctx context.Context, bucketName, objectName string, opts ...PutOption) (string, error) {
	// 验证参数
	if err := validateBucketName(bucketName); err != nil {
		return "", err
	}
	if err := validateObjectName(objectName); err != nil {
		return "", err
	}

	// 应用选项
	options := applyPutOptions(opts)

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName:   bucketName,
		ObjectName:   objectName,
		QueryValues:  url.Values{},
		CustomHeader: make(http.Header),
	}

	// 设置 uploads 查询参数
	meta.QueryValues.Set("uploads", "")

	// 设置 Content-Type
	if options.ContentType != "" {
		meta.CustomHeader.Set("Content-Type", options.ContentType)
	}

	// 设置 Content-Encoding
	if options.ContentEncoding != "" {
		meta.CustomHeader.Set("Content-Encoding", options.ContentEncoding)
	}

	// 设置 Content-Disposition
	if options.ContentDisposition != "" {
		meta.CustomHeader.Set("Content-Disposition", options.ContentDisposition)
	}

	// 设置存储类
	if options.StorageClass != "" {
		meta.CustomHeader.Set("x-amz-storage-class", options.StorageClass)
	}

	// 设置用户元数据
	if options.UserMetadata != nil {
		for k, v := range options.UserMetadata {
			meta.CustomHeader.Set("x-amz-meta-"+k, v)
		}
	}

	// 设置用户标签
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

	// 合并自定义头
	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	// 创建 POST 请求
	req := core.NewRequest(ctx, http.MethodPost, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return "", err
	}
	defer closeResponse(resp)

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return "", parseErrorResponse(resp, bucketName, objectName)
	}

	// 解析响应
	var result initiateMultipartUploadResult
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode initiate multipart upload response: %w", err)
	}

	return result.UploadID, nil
}

// UploadPart 上传分片
func (s *objectService) UploadPart(ctx context.Context, bucketName, objectName, uploadID string, partNumber int, reader io.Reader, partSize int64, opts ...PutOption) (types.ObjectPart, error) {
	// 验证参数
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

	// 应用选项
	options := applyPutOptions(opts)

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName:    bucketName,
		ObjectName:    objectName,
		ContentBody:   reader,
		ContentLength: partSize,
		QueryValues:   url.Values{},
		CustomHeader:  make(http.Header),
	}

	// 设置查询参数
	meta.QueryValues.Set("uploadId", uploadID)
	meta.QueryValues.Set("partNumber", strconv.Itoa(partNumber))

	// 合并自定义头
	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	// 创建 PUT 请求
	req := core.NewRequest(ctx, http.MethodPut, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return types.ObjectPart{}, err
	}
	defer closeResponse(resp)

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return types.ObjectPart{}, parseErrorResponse(resp, bucketName, objectName)
	}

	// 构建分片信息
	part := types.ObjectPart{
		PartNumber: partNumber,
		ETag:       trimETag(resp.Header.Get("ETag")),
		Size:       partSize,
	}

	// 解析校验和
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

// CompleteMultipartUpload 完成分片上传
func (s *objectService) CompleteMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string, parts []types.ObjectPart, opts ...PutOption) (types.UploadInfo, error) {
	// 验证参数
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

	// 应用选项
	options := applyPutOptions(opts)

	// 构建完成分片上传请求
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

	// XML 编码
	xmlData, err := xml.Marshal(completeUpload)
	if err != nil {
		return types.UploadInfo{}, fmt.Errorf("failed to marshal complete multipart upload: %w", err)
	}

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName:    bucketName,
		ObjectName:    objectName,
		ContentBody:   nil, // 将在后面设置
		ContentLength: int64(len(xmlData)),
		QueryValues:   url.Values{},
		CustomHeader:  make(http.Header),
	}

	// 设置查询参数
	meta.QueryValues.Set("uploadId", uploadID)

	// 设置 Content-Type
	meta.CustomHeader.Set("Content-Type", "application/xml")

	// 合并自定义头
	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	// 创建 POST 请求（使用 XML 数据作为 body）
	meta.ContentBody = bytes.NewReader(xmlData)

	req := core.NewRequest(ctx, http.MethodPost, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return types.UploadInfo{}, err
	}
	defer closeResponse(resp)

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return types.UploadInfo{}, parseErrorResponse(resp, bucketName, objectName)
	}

	// 解析响应
	var result completeMultipartUploadResult
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&result); err != nil {
		return types.UploadInfo{}, fmt.Errorf("failed to decode complete multipart upload response: %w", err)
	}

	// 构建上传信息
	uploadInfo := types.UploadInfo{
		Bucket:       result.Bucket,
		Key:          result.Key,
		ETag:         trimETag(result.ETag),
		VersionID:    resp.Header.Get("x-amz-version-id"),
		LastModified: result.Modified,
	}

	// 计算总大小
	var totalSize int64
	for _, part := range parts {
		totalSize += part.Size
	}
	uploadInfo.Size = totalSize

	// 解析校验和
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

// AbortMultipartUpload 取消分片上传
func (s *objectService) AbortMultipartUpload(ctx context.Context, bucketName, objectName, uploadID string) error {
	// 验证参数
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if err := validateObjectName(objectName); err != nil {
		return err
	}
	if uploadID == "" {
		return fmt.Errorf("upload ID cannot be empty")
	}

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName:  bucketName,
		ObjectName:  objectName,
		QueryValues: url.Values{},
	}

	// 设置查询参数
	meta.QueryValues.Set("uploadId", uploadID)

	// 创建 DELETE 请求
	req := core.NewRequest(ctx, http.MethodDelete, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	// 检查响应（204 No Content）
	if resp.StatusCode != http.StatusNoContent {
		return parseErrorResponse(resp, bucketName, objectName)
	}

	return nil
}
