// Package bucket bucket/service.go
package bucket

import (
	"context"

	"github.com/Scorpio69t/rustfs-go/types"
)

// Service Bucket 服务接口
type Service interface {
	// Create 创建桶
	Create(ctx context.Context, bucketName string, opts ...CreateOption) error

	// Delete 删除桶
	Delete(ctx context.Context, bucketName string, opts ...DeleteOption) error

	// Exists 检查桶是否存在
	Exists(ctx context.Context, bucketName string) (bool, error)

	// List 列出所有桶
	List(ctx context.Context) ([]types.BucketInfo, error)

	// GetLocation 获取桶位置
	GetLocation(ctx context.Context, bucketName string) (string, error)
}

// CreateOption 创建桶选项函数
type CreateOption func(*CreateOptions)

// CreateOptions 创建桶选项
type CreateOptions struct {
	// Region 区域
	Region string

	// ObjectLocking 启用对象锁定
	ObjectLocking bool

	// ForceCreate 强制创建（RustFS 扩展）
	ForceCreate bool
}

// DeleteOption 删除桶选项函数
type DeleteOption func(*DeleteOptions)

// DeleteOptions 删除桶选项
type DeleteOptions struct {
	// ForceDelete 强制删除（即使桶不为空）
	ForceDelete bool
}

// WithRegion 设置区域
func WithRegion(region string) CreateOption {
	return func(opts *CreateOptions) {
		opts.Region = region
	}
}

// WithObjectLocking 启用对象锁定
func WithObjectLocking(enabled bool) CreateOption {
	return func(opts *CreateOptions) {
		opts.ObjectLocking = enabled
	}
}

// WithForceCreate 强制创建
func WithForceCreate(force bool) CreateOption {
	return func(opts *CreateOptions) {
		opts.ForceCreate = force
	}
}

// WithForceDelete 强制删除
func WithForceDelete(force bool) DeleteOption {
	return func(opts *DeleteOptions) {
		opts.ForceDelete = force
	}
}
