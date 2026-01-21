// Package bucket bucket/service.go
package bucket

import (
	"context"

	"github.com/Scorpio69t/rustfs-go/pkg/sse"
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

	// SetPolicy sets bucket policy JSON
	SetPolicy(ctx context.Context, bucketName, policy string) error

	// GetPolicy retrieves bucket policy JSON
	GetPolicy(ctx context.Context, bucketName string) (string, error)

	// DeletePolicy removes bucket policy
	DeletePolicy(ctx context.Context, bucketName string) error

	// SetLifecycle sets bucket lifecycle configuration (XML)
	SetLifecycle(ctx context.Context, bucketName string, config []byte) error

	// GetLifecycle retrieves bucket lifecycle configuration (XML)
	GetLifecycle(ctx context.Context, bucketName string) ([]byte, error)

	// DeleteLifecycle removes bucket lifecycle configuration
	DeleteLifecycle(ctx context.Context, bucketName string) error

	// SetVersioning sets bucket versioning configuration
	SetVersioning(ctx context.Context, bucketName string, cfg types.VersioningConfig) error

	// GetVersioning retrieves bucket versioning configuration
	GetVersioning(ctx context.Context, bucketName string) (types.VersioningConfig, error)

	// SetReplication sets bucket cross-region replication configuration
	SetReplication(ctx context.Context, bucketName string, config []byte) error

	// GetReplication retrieves bucket replication configuration
	GetReplication(ctx context.Context, bucketName string) ([]byte, error)

	// DeleteReplication removes bucket replication configuration
	DeleteReplication(ctx context.Context, bucketName string) error

	// SetNotification sets bucket event notification configuration
	SetNotification(ctx context.Context, bucketName string, config []byte) error

	// GetNotification retrieves bucket event notification configuration
	GetNotification(ctx context.Context, bucketName string) ([]byte, error)

	// DeleteNotification removes bucket event notification configuration
	DeleteNotification(ctx context.Context, bucketName string) error

	// SetLogging sets bucket access logging configuration
	SetLogging(ctx context.Context, bucketName string, config []byte) error

	// GetLogging retrieves bucket access logging configuration
	GetLogging(ctx context.Context, bucketName string) ([]byte, error)

	// DeleteLogging removes bucket access logging configuration
	DeleteLogging(ctx context.Context, bucketName string) error

	// SetEncryption sets bucket default encryption configuration
	SetEncryption(ctx context.Context, bucketName string, config sse.Configuration) error

	// GetEncryption retrieves bucket default encryption configuration
	GetEncryption(ctx context.Context, bucketName string) (sse.Configuration, error)

	// DeleteEncryption removes bucket default encryption configuration
	DeleteEncryption(ctx context.Context, bucketName string) error
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
