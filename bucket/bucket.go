// Package bucket bucket/bucket.go
package bucket

import (
	"github.com/Scorpio69t/rustfs-go/internal/cache"
	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// bucketService Bucket 服务实现
type bucketService struct {
	executor      *core.Executor
	locationCache *cache.LocationCache
}

// NewService 创建 Bucket 服务
func NewService(executor *core.Executor, locationCache *cache.LocationCache) Service {
	return &bucketService{
		executor:      executor,
		locationCache: locationCache,
	}
}

// applyCreateOptions 应用创建选项
func applyCreateOptions(opts []CreateOption) CreateOptions {
	options := CreateOptions{
		Region: "us-east-1", // 默认区域
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyDeleteOptions 应用删除选项
func applyDeleteOptions(opts []DeleteOption) DeleteOptions {
	options := DeleteOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// validateBucketName 验证桶名
func validateBucketName(bucketName string) error {
	if bucketName == "" {
		return ErrInvalidBucketName
	}
	// 桶名长度检查
	if len(bucketName) < 3 || len(bucketName) > 63 {
		return ErrInvalidBucketName
	}
	return nil
}
