// Package object object/object.go
package object

import (
	"github.com/Scorpio69t/rustfs-go/internal/cache"
	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// objectService implements Object service
type objectService struct {
	executor      *core.Executor
	locationCache *cache.LocationCache
}

// NewService creates a new Object service
func NewService(executor *core.Executor, locationCache *cache.LocationCache) Service {
	return &objectService{
		executor:      executor,
		locationCache: locationCache,
	}
}

// Put, Get, Stat, Delete, List, Copy are implemented in separate files
// - put.go: Put method
// - get.go: Get method
// - stat.go: Stat method
// - delete.go: Delete method
// - list.go: List method
// - copy.go: Copy method
// - multipart.go: InitiateMultipartUpload, UploadPart, CompleteMultipartUpload, AbortMultipartUpload methods
// - compose.go: Compose method
// - append.go: Append method
// - select.go: Select method

// applyPutOptions applies upload options
func applyPutOptions(opts []PutOption) PutOptions {
	options := PutOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyGetOptions applies download options
func applyGetOptions(opts []GetOption) GetOptions {
	options := GetOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyStatOptions applies stat options
func applyStatOptions(opts []StatOption) StatOptions {
	options := StatOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyDeleteOptions applies delete options
func applyDeleteOptions(opts []DeleteOption) DeleteOptions {
	options := DeleteOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyListOptions applies list options
func applyListOptions(opts []ListOption) ListOptions {
	options := ListOptions{
		MaxKeys: 1000, // default max keys
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyCopyOptions applies copy options
func applyCopyOptions(opts []CopyOption) CopyOptions {
	options := CopyOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyLegalHoldOptions applies legal hold options.
func applyLegalHoldOptions(opts []LegalHoldOption) LegalHoldOptions {
	options := LegalHoldOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyRetentionOptions applies retention options.
func applyRetentionOptions(opts []RetentionOption) RetentionOptions {
	options := RetentionOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// validateBucketName validates bucket name
func validateBucketName(bucketName string) error {
	if bucketName == "" {
		return ErrInvalidBucketName
	}
	if len(bucketName) < 3 || len(bucketName) > 63 {
		return ErrInvalidBucketName
	}
	return nil
}

// validateObjectName validates object name
func validateObjectName(objectName string) error {
	if objectName == "" {
		return ErrInvalidObjectName
	}
	return nil
}
