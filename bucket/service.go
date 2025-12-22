// Package bucket bucket/service.go
package bucket

import (
	"context"

	"github.com/Scorpio69t/rustfs-go/types"
)

// Service Bucket interface
type Service interface {
	// Create bucket
	Create(ctx context.Context, bucketName string, opts ...CreateOption) error

	// Delete bucket
	Delete(ctx context.Context, bucketName string, opts ...DeleteOption) error

	// Exists check if bucket exists
	Exists(ctx context.Context, bucketName string) (bool, error)

	// List buckets
	List(ctx context.Context) ([]types.BucketInfo, error)

	// GetLocation get bucket location/region
	GetLocation(ctx context.Context, bucketName string) (string, error)
}

// CreateOption create bucket options function
type CreateOption func(*CreateOptions)

// CreateOptions create bucket options
type CreateOptions struct {
	// Region
	Region string

	// ObjectLocking enable object locking
	ObjectLocking bool

	// ForceCreate force create (overwrite if exists)
	ForceCreate bool
}

// DeleteOption delete bucket options function
type DeleteOption func(*DeleteOptions)

// DeleteOptions delete bucket options
type DeleteOptions struct {
	// ForceDelete force delete (delete even if not empty)
	ForceDelete bool
}

// WithRegion set bucket region
func WithRegion(region string) CreateOption {
	return func(opts *CreateOptions) {
		opts.Region = region
	}
}

// WithObjectLocking set object locking
func WithObjectLocking(enabled bool) CreateOption {
	return func(opts *CreateOptions) {
		opts.ObjectLocking = enabled
	}
}

// WithForceCreate set force create
func WithForceCreate(force bool) CreateOption {
	return func(opts *CreateOptions) {
		opts.ForceCreate = force
	}
}

// WithForceDelete set force delete
func WithForceDelete(force bool) DeleteOption {
	return func(opts *DeleteOptions) {
		opts.ForceDelete = force
	}
}
