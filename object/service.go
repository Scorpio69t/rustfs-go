// Package object object/service.go
package object

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Scorpio69t/rustfs-go/pkg/acl"
	"github.com/Scorpio69t/rustfs-go/pkg/objectlock"
	s3select "github.com/Scorpio69t/rustfs-go/pkg/select"
	"github.com/Scorpio69t/rustfs-go/types"
)

// Service Object service interface
type Service interface {
	// Put uploads an object
	Put(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts ...PutOption) (types.UploadInfo, error)

	// Get downloads an object
	Get(ctx context.Context, bucketName, objectName string, opts ...GetOption) (io.ReadCloser, types.ObjectInfo, error)

	// FPut uploads a file from a local path
	FPut(ctx context.Context, bucketName, objectName, filePath string, opts ...PutOption) (types.UploadInfo, error)

	// FGet downloads an object to a local file path
	FGet(ctx context.Context, bucketName, objectName, filePath string, opts ...GetOption) (types.ObjectInfo, error)

	// Stat retrieves object info
	Stat(ctx context.Context, bucketName, objectName string, opts ...StatOption) (types.ObjectInfo, error)

	// Delete removes an object
	Delete(ctx context.Context, bucketName, objectName string, opts ...DeleteOption) error

	// List lists objects
	List(ctx context.Context, bucketName string, opts ...ListOption) <-chan types.ObjectInfo

	// ListVersions lists object versions and delete markers
	ListVersions(ctx context.Context, bucketName string, opts ...ListOption) <-chan types.ObjectInfo

	// Copy copies an object
	Copy(ctx context.Context, destBucket, destObject, srcBucket, srcObject string, opts ...CopyOption) (types.CopyInfo, error)

	// Compose creates an object by composing source objects
	Compose(ctx context.Context, dst DestinationInfo, sources []SourceInfo, opts ...PutOption) (types.UploadInfo, error)

	// Append appends data to an object at a specific offset
	Append(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, offset int64, opts ...PutOption) (types.UploadInfo, error)

	// Select queries object content using S3 Select
	Select(ctx context.Context, bucketName, objectName string, opts s3select.Options) (*s3select.Results, error)

	// PresignGet creates a presigned GET URL with optional signed headers
	PresignGet(ctx context.Context, bucketName, objectName string, expires time.Duration, reqParams url.Values, opts ...PresignOption) (*url.URL, http.Header, error)

	// PresignPut creates a presigned PUT URL with optional signed headers
	PresignPut(ctx context.Context, bucketName, objectName string, expires time.Duration, reqParams url.Values, opts ...PresignOption) (*url.URL, http.Header, error)

	// SetTagging sets object tags
	SetTagging(ctx context.Context, bucketName, objectName string, tags map[string]string) error

	// GetTagging retrieves object tags
	GetTagging(ctx context.Context, bucketName, objectName string) (map[string]string, error)

	// DeleteTagging deletes object tags
	DeleteTagging(ctx context.Context, bucketName, objectName string) error

	// GetACL retrieves object ACL
	GetACL(ctx context.Context, bucketName, objectName string) (acl.ACL, error)

	// SetACL sets object ACL
	SetACL(ctx context.Context, bucketName, objectName string, policy acl.ACL) error

	// SetLegalHold sets legal hold status for an object
	SetLegalHold(ctx context.Context, bucketName, objectName string, hold objectlock.LegalHoldStatus, opts ...LegalHoldOption) error

	// GetLegalHold retrieves legal hold status for an object
	GetLegalHold(ctx context.Context, bucketName, objectName string, opts ...LegalHoldOption) (objectlock.LegalHoldStatus, error)

	// SetRetention sets retention mode and retain-until date for an object
	SetRetention(ctx context.Context, bucketName, objectName string, mode objectlock.RetentionMode, retainUntil time.Time, opts ...RetentionOption) error

	// GetRetention retrieves retention configuration for an object
	GetRetention(ctx context.Context, bucketName, objectName string, opts ...RetentionOption) (objectlock.RetentionMode, time.Time, error)
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

// PresignOption applies presign options
type PresignOption func(*PresignOptions)

// LegalHoldOption applies legal hold options
type LegalHoldOption func(*LegalHoldOptions)

// RetentionOption applies retention options
type RetentionOption func(*RetentionOptions)
