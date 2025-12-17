// Package bucket bucket/delete.go
package bucket

import (
	"context"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// Delete 删除桶
func (s *bucketService) Delete(ctx context.Context, bucketName string, opts ...DeleteOption) error {
	// 验证桶名
	if err := validateBucketName(bucketName); err != nil {
		return err
	}

	// 应用选项
	options := applyDeleteOptions(opts)

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName:   bucketName,
		CustomHeader: make(http.Header),
	}

	// 设置强制删除头（如果支持）
	if options.ForceDelete {
		meta.CustomHeader.Set("x-minio-force-delete", "true")
	}

	// 创建请求
	req := core.NewRequest(ctx, http.MethodDelete, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		return err
	}
	defer closeResponse(resp)

	// 检查响应
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return parseErrorResponse(resp, bucketName, "")
	}

	// 成功后从缓存中删除
	if s.locationCache != nil {
		s.locationCache.Delete(bucketName)
	}

	return nil
}
