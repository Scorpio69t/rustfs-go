// Package object object/get.go
package object

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// Get 下载对象（实现）
func (s *objectService) Get(ctx context.Context, bucketName, objectName string, opts ...GetOption) (io.ReadCloser, types.ObjectInfo, error) {
	// 验证参数
	if err := validateBucketName(bucketName); err != nil {
		return nil, types.ObjectInfo{}, err
	}
	if err := validateObjectName(objectName); err != nil {
		return nil, types.ObjectInfo{}, err
	}

	// 应用选项
	options := applyGetOptions(opts)

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName:   bucketName,
		ObjectName:   objectName,
		CustomHeader: make(http.Header),
	}

	// 设置 Range 头
	if options.SetRange {
		rangeHeader := "bytes=" + strconv.FormatInt(options.RangeStart, 10) + "-"
		if options.RangeEnd > 0 {
			rangeHeader += strconv.FormatInt(options.RangeEnd, 10)
		}
		meta.CustomHeader.Set("Range", rangeHeader)
	}

	// 设置条件匹配头
	if options.MatchETag != "" {
		meta.CustomHeader.Set("If-Match", options.MatchETag)
	}
	if options.NotMatchETag != "" {
		meta.CustomHeader.Set("If-None-Match", options.NotMatchETag)
	}
	if !options.MatchModified.IsZero() {
		meta.CustomHeader.Set("If-Modified-Since", options.MatchModified.Format(http.TimeFormat))
	}
	if !options.NotModified.IsZero() {
		meta.CustomHeader.Set("If-Unmodified-Since", options.NotModified.Format(http.TimeFormat))
	}

	// 添加版本 ID 查询参数
	if options.VersionID != "" {
		meta.QueryValues = url.Values{}
		meta.QueryValues.Set("versionId", options.VersionID)
	}

	// 合并自定义头
	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	// 创建 GET 请求
	req := core.NewRequest(ctx, http.MethodGet, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return nil, types.ObjectInfo{}, err
	}

	// 检查响应（200 OK 或 206 Partial Content）
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		closeResponse(resp)
		return nil, types.ObjectInfo{}, parseErrorResponse(resp, bucketName, objectName)
	}

	// 解析对象信息
	parser := core.NewResponseParser()
	objectInfo, err := parser.ParseObjectInfo(resp, bucketName, objectName)
	if err != nil {
		closeResponse(resp)
		return nil, types.ObjectInfo{}, err
	}

	// 返回响应体和对象信息
	// 注意：调用者负责关闭 Body
	return resp.Body, objectInfo, nil
}
