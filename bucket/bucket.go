// Package bucket bucket/bucket.go
package bucket

import (
	"github.com/Scorpio69t/rustfs-go/internal/cache"
	"github.com/Scorpio69t/rustfs-go/internal/core"
)

// bucketService Bucket Service struct
type bucketService struct {
	executor      *core.Executor
	locationCache *cache.LocationCache
}

// NewService Create a new Bucket Service
func NewService(executor *core.Executor, locationCache *cache.LocationCache) Service {
	return &bucketService{
		executor:      executor,
		locationCache: locationCache,
	}
}

// applyCreateOptions apply to create options
func applyCreateOptions(opts []CreateOption) CreateOptions {
	options := CreateOptions{
		Region: "us-east-1", // default region
	}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// applyDeleteOptions apply delete options
func applyDeleteOptions(opts []DeleteOption) DeleteOptions {
	options := DeleteOptions{}
	for _, opt := range opts {
		opt(&options)
	}
	return options
}

// validateBucketName validate bucket name
func validateBucketName(bucketName string) error {
	if bucketName == "" {
		return ErrInvalidBucketName
	}

	// check length
	if len(bucketName) < 3 || len(bucketName) > 63 {
		return ErrInvalidBucketName
	}
	return nil
}
