// Package object object/service.go
package object

import (
	"context"
	"io"

	"github.com/Scorpio69t/rustfs-go/types"
)

// Service Object service interface
type Service interface {
	// Put uploads an object
	Put(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts ...PutOption) (types.UploadInfo, error)

	// Get downloads an object
	Get(ctx context.Context, bucketName, objectName string, opts ...GetOption) (io.ReadCloser, types.ObjectInfo, error)

	// Stat retrieves object info
	Stat(ctx context.Context, bucketName, objectName string, opts ...StatOption) (types.ObjectInfo, error)

	// Delete removes an object
	Delete(ctx context.Context, bucketName, objectName string, opts ...DeleteOption) error

	// List lists objects
	List(ctx context.Context, bucketName string, opts ...ListOption) <-chan types.ObjectInfo

	// Copy copies an object
	Copy(ctx context.Context, destBucket, destObject, srcBucket, srcObject string, opts ...CopyOption) (types.CopyInfo, error)
}

// PutOption applies upload option
type PutOption func(*PutOptions)

// GetOption applies download option
type GetOption func(*GetOptions)

// StatOption applies object info option
type StatOption func(*StatOptions)

// DeleteOption applies delete option
type DeleteOption func(*DeleteOptions)

// ListOption applies list option
type ListOption func(*ListOptions)

// CopyOption applies copy option
type CopyOption func(*CopyOptions)
