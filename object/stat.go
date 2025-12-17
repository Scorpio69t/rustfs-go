// Package object object/stat.go
package object

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
	"github.com/Scorpio69t/rustfs-go/types"
)

// Stat 获取对象信息（实现）
func (s *objectService) Stat(ctx context.Context, bucketName, objectName string, opts ...StatOption) (types.ObjectInfo, error) {
	// 验证参数
	if err := validateBucketName(bucketName); err != nil {
		return types.ObjectInfo{}, err
	}
	if err := validateObjectName(objectName); err != nil {
		return types.ObjectInfo{}, err
	}

	// 应用选项
	options := applyStatOptions(opts)

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName:   bucketName,
		ObjectName:   objectName,
		CustomHeader: options.CustomHeaders,
	}

	// 添加版本 ID 查询参数
	if options.VersionID != "" {
		meta.QueryValues = url.Values{}
		meta.QueryValues.Set("versionId", options.VersionID)
	}

	// 创建 HEAD 请求
	req := core.NewRequest(ctx, http.MethodHead, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return types.ObjectInfo{}, err
	}
	defer closeResponse(resp)

	// 检查响应
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return types.ObjectInfo{}, parseErrorResponse(resp, bucketName, objectName)
	}

	// 解析对象信息
	parser := core.NewResponseParser()
	return parser.ParseObjectInfo(resp, bucketName, objectName)
}
