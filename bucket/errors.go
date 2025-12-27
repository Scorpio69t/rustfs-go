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

	// ErrInvalidVersioningStatus invalid versioning status
	ErrInvalidVersioningStatus = errors.New("invalid versioning status, must be Enabled or Suspended")

	// ErrEmptyBucketConfig invalid empty bucket configuration payload
	ErrEmptyBucketConfig = errors.New("configuration payload cannot be empty")
)
