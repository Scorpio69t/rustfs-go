// Package object object/errors.go
package object

import "errors"

var (
	// ErrInvalidBucketName invalid bucket name
	ErrInvalidBucketName = errors.New("invalid bucket name")

	// ErrInvalidObjectName invalid object name
	ErrInvalidObjectName = errors.New("invalid object name")

	// ErrObjectNotFound object not found
	ErrObjectNotFound = errors.New("object not found")

	// ErrNotImplemented feature not implemented
	ErrNotImplemented = errors.New("not implemented yet")
)
