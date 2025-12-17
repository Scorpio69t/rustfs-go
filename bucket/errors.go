// Package bucket bucket/errors.go
package bucket

import "errors"

var (
	// ErrInvalidBucketName 无效的桶名
	ErrInvalidBucketName = errors.New("invalid bucket name")

	// ErrBucketNotFound 桶不存在
	ErrBucketNotFound = errors.New("bucket not found")

	// ErrBucketAlreadyExists 桶已存在
	ErrBucketAlreadyExists = errors.New("bucket already exists")

	// ErrBucketNotEmpty 桶不为空
	ErrBucketNotEmpty = errors.New("bucket not empty")
)
