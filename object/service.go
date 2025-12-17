// Package object object/service.go
package object

import (
	"context"
	"io"

	"github.com/Scorpio69t/rustfs-go/types"
)

// Service Object 服务接口
type Service interface {
	// Put 上传对象
	Put(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts ...PutOption) (types.UploadInfo, error)

	// Get 下载对象
	Get(ctx context.Context, bucketName, objectName string, opts ...GetOption) (io.ReadCloser, types.ObjectInfo, error)

	// Stat 获取对象信息
	Stat(ctx context.Context, bucketName, objectName string, opts ...StatOption) (types.ObjectInfo, error)

	// Delete 删除对象
	Delete(ctx context.Context, bucketName, objectName string, opts ...DeleteOption) error

	// List 列出对象
	List(ctx context.Context, bucketName string, opts ...ListOption) <-chan types.ObjectInfo

	// Copy 复制对象
	Copy(ctx context.Context, destBucket, destObject, srcBucket, srcObject string, opts ...CopyOption) (types.CopyInfo, error)
}

// PutOption 上传选项函数
type PutOption func(*PutOptions)

// GetOption 下载选项函数
type GetOption func(*GetOptions)

// StatOption 获取对象信息选项函数
type StatOption func(*StatOptions)

// DeleteOption 删除选项函数
type DeleteOption func(*DeleteOptions)

// ListOption 列出对象选项函数
type ListOption func(*ListOptions)

// CopyOption 复制选项函数
type CopyOption func(*CopyOptions)
