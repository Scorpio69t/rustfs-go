// Package object object/delete.go
package object

import (
	"context"
	"net/http"
	"net/url"

	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// Delete 删除对象（实现）
func (s *objectService) Delete(ctx context.Context, bucketName, objectName string, opts ...DeleteOption) error {
	// 验证参数
	if err := validateBucketName(bucketName); err != nil {
		return err
	}
	if err := validateObjectName(objectName); err != nil {
		return err
	}

	// 应用选项
	options := applyDeleteOptions(opts)

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName:   bucketName,
		ObjectName:   objectName,
		CustomHeader: make(http.Header),
	}

	// 添加版本 ID 查询参数
	if options.VersionID != "" {
		meta.QueryValues = url.Values{}
		meta.QueryValues.Set("versionId", options.VersionID)
	}

	// 设置强制删除头（如果支持）
	if options.ForceDelete {
		meta.CustomHeader.Set("x-minio-force-delete", "true")
	}

	// 合并自定义头
	if options.CustomHeaders != nil {
		for k, v := range options.CustomHeaders {
			meta.CustomHeader[k] = v
		}
	}

	// 创建 DELETE 请求
	req := core.NewRequest(ctx, http.MethodDelete, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	// 检查响应（204 No Content 或 200 OK）
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp, bucketName, objectName)
	}

	return nil
}
