// Package object object/object.go
package object

import (
	"github.com/Scorpio69t/rustfs-go/internal/cache"
	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// objectService Object 服务实现
type objectService struct {
	executor      *core.Executor
	locationCache *cache.LocationCache
}

// NewService 创建 Object 服务
func NewService(executor *core.Executor, locationCache *cache.LocationCache) Service {
	return &objectService{
		executor:      executor,
		locationCache: locationCache,
	}
}

// Put, Get, Stat, Delete, List, Copy 方法已在独立文件中实现
// - put.go: Put 方法
// - get.go: Get 方法
// - stat.go: Stat 方法
// - delete.go: Delete 方法
// - list.go: List 方法
// - copy.go: Copy 方法
// - multipart.go: InitiateMultipartUpload, UploadPart, CompleteMultipartUpload, AbortMultipartUpload 方法

// applyPutOptions 应用上传选项
func applyPutOptions(opts []PutOption) PutOptions {
	options := PutOptions{
		ContentType: "application/octet-stream", // 默认内容类型
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyGetOptions 应用下载选项
func applyGetOptions(opts []GetOption) GetOptions {
	options := GetOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyStatOptions 应用获取对象信息选项
func applyStatOptions(opts []StatOption) StatOptions {
	options := StatOptions{}
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

// applyListOptions 应用列出对象选项
func applyListOptions(opts []ListOption) ListOptions {
	options := ListOptions{
		MaxKeys: 1000, // 默认最大键数
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyCopyOptions 应用复制选项
func applyCopyOptions(opts []CopyOption) CopyOptions {
	options := CopyOptions{}
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
	if len(bucketName) < 3 || len(bucketName) > 63 {
		return ErrInvalidBucketName
	}
	return nil
}

// validateObjectName 验证对象名
func validateObjectName(objectName string) error {
	if objectName == "" {
		return ErrInvalidObjectName
	}
	return nil
}
