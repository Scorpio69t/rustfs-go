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

// copyObjectResult 复制对象结果（XML 响应）
type copyObjectResult struct {
	XMLName      xml.Name  `xml:"CopyObjectResult"`
	ETag         string    `xml:"ETag"`
	LastModified time.Time `xml:"LastModified"`
}

// Copy 复制对象（实现）
func (s *objectService) Copy(ctx context.Context, destBucket, destObject, sourceBucket, sourceObject string, opts ...CopyOption) (types.CopyInfo, error) {
	// 验证参数
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

	// 应用选项
	options := applyCopyOptions(opts)

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName:   destBucket,
		ObjectName:   destObject,
		CustomHeader: make(http.Header),
	}

	// 设置复制源头
	copySource := fmt.Sprintf("%s/%s", sourceBucket, sourceObject)
	if options.SourceVersionID != "" {
		copySource += "?versionId=" + options.SourceVersionID
	}
	meta.CustomHeader.Set("x-amz-copy-source", copySource)

	// 设置元数据指令
	if options.ReplaceMetadata {
		meta.CustomHeader.Set("x-amz-metadata-directive", "REPLACE")

		// 设置新的元数据
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

		// 设置用户元数据
		if options.UserMetadata != nil {
			for k, v := range options.UserMetadata {
				meta.CustomHeader.Set("x-amz-meta-"+k, v)
			}
		}
	} else {
		meta.CustomHeader.Set("x-amz-metadata-directive", "COPY")
	}

	// 设置标签指令
	if options.ReplaceTagging {
		meta.CustomHeader.Set("x-amz-tagging-directive", "REPLACE")

		// 设置新的标签
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

	// 设置条件匹配头（源对象）
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

	// 设置存储类
	if options.StorageClass != "" {
		meta.CustomHeader.Set("x-amz-storage-class", options.StorageClass)
	}

	// 合并自定义头
	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	// 创建 PUT 请求（复制使用 PUT 方法）
	req := core.NewRequest(ctx, http.MethodPut, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return types.CopyInfo{}, err
	}
	defer closeResponse(resp)

	// 检查响应
	if resp.StatusCode != http.StatusOK {
		return types.CopyInfo{}, parseErrorResponse(resp, destBucket, destObject)
	}

	// 解析复制结果
	var cpResult copyObjectResult
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&cpResult); err != nil {
		return types.CopyInfo{}, fmt.Errorf("failed to decode copy object response: %w", err)
	}

	// 构建复制信息
	copyInfo := types.CopyInfo{
		Bucket:          destBucket,
		Key:             destObject,
		ETag:            trimETag(cpResult.ETag),
		LastModified:    cpResult.LastModified,
		VersionID:       resp.Header.Get("x-amz-version-id"),
		SourceVersionID: options.SourceVersionID,
	}

	// 解析校验和
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
