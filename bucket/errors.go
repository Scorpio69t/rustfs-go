// Package bucket bucket/errors.go
package bucket

import "errors"

var (
	// ErrInvalidBucketName invalid bucket name
	ErrInvalidBucketName = errors.New("invalid bucket name")

	// ErrBucketNotFound not found bucket
	ErrBucketNotFound = errors.New("bucket not found")

	// ErrBucketAlreadyExists bucket already exists
	ErrBucketAlreadyExists = errors.New("bucket already exists")

	// ErrBucketNotEmpty bucket not empty
	ErrBucketNotEmpty = errors.New("bucket not empty")
)
