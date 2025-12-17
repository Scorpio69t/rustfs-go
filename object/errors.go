// Package object object/errors.go
package object

import "errors"

var (
	// ErrInvalidBucketName 无效的桶名
	ErrInvalidBucketName = errors.New("invalid bucket name")

	// ErrInvalidObjectName 无效的对象名
	ErrInvalidObjectName = errors.New("invalid object name")

	// ErrObjectNotFound 对象不存在
	ErrObjectNotFound = errors.New("object not found")

	// ErrNotImplemented 功能未实现
	ErrNotImplemented = errors.New("not implemented yet")
)
