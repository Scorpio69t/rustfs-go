// Package bucket bucket/exists.go
package bucket

import (
	"context"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/errors"
	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// Exists 检查桶是否存在
func (s *bucketService) Exists(ctx context.Context, bucketName string) (bool, error) {
	// 验证桶名
	if err := validateBucketName(bucketName); err != nil {
		return false, err
	}

	// 构建请求元数据
	meta := core.RequestMetadata{
		BucketName: bucketName,
	}

	// 创建 HEAD 请求
	req := core.NewRequest(ctx, http.MethodHead, meta)

	// 执行请求
	resp, err := s.executor.Execute(ctx, req)
	if err != nil {
		// 检查是否为 NoSuchBucket 错误
		if apiErr, ok := err.(*errors.APIError); ok {
			if apiErr.Code() == errors.ErrCodeNoSuchBucket {
				return false, nil
			}
		}
		return false, err
	}
	defer closeResponse(resp)

	// 检查响应状态
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode != http.StatusOK {
		err := parseErrorResponse(resp, bucketName, "")
		// 检查是否为 NoSuchBucket 错误
		if apiErr, ok := err.(*errors.APIError); ok {
			if apiErr.Code() == errors.ErrCodeNoSuchBucket {
				return false, nil
			}
		}
		return false, err
	}

	return true, nil
}
