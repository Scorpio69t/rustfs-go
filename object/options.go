// Package object object/options.go
package object

import (
	"net/http"
	"net/url"
	"time"
)

// PutOptions controls object upload behavior
type PutOptions struct {
	// Tracks whether caller explicitly set Content-Type
	contentTypeSet bool

	// Content-Type
	ContentType string

	// Content-Encoding
	ContentEncoding string

	// Content-Disposition
	ContentDisposition string

	// Content-Language
	ContentLanguage string

	// Cache-Control
	CacheControl string

	// Expiration time
	Expires time.Time

	// User metadata
	UserMetadata map[string]string

	// User tags
	UserTags map[string]string

	// Storage class
	StorageClass string

	// Custom headers
	CustomHeaders http.Header

	// Whether to send Content-MD5
	SendContentMD5 bool

	// Disable Content-SHA256
	DisableContentSHA256 bool

	// SSE-S3 / SSE-C options
	SSECustomerAlgorithm string
	SSECustomerKey       string
	SSECustomerKeyMD5    string

	// Multipart part size
	PartSize uint64

	// Number of concurrent uploads
	NumThreads uint
}

// PresignOptions controls presigned URL generation
type PresignOptions struct {
	Headers     http.Header
	QueryValues url.Values
}

// GetOptions controls object download behavior
type GetOptions struct {
	// Range header
	RangeStart int64
	RangeEnd   int64
	SetRange   bool

	// Version ID
	VersionID string

	// Conditional headers
	MatchETag     string
	NotMatchETag  string
	MatchModified time.Time
	NotModified   time.Time

	// Custom headers
	CustomHeaders http.Header

	// SSE-C headers for encrypted objects
	SSECustomerAlgorithm string
	SSECustomerKey       string
	SSECustomerKeyMD5    string
}

// StatOptions controls stat/metadata retrieval
type StatOptions struct {
	// Version ID
	VersionID string

	// Custom headers
	CustomHeaders http.Header
}

// DeleteOptions controls object deletion
type DeleteOptions struct {
	// Version ID
	VersionID string

	// Force delete when possible
	ForceDelete bool

	// Custom headers
	CustomHeaders http.Header
}

// ListOptions controls object listing
type ListOptions struct {
	// Prefix filter
	Prefix string

	// Recursive listing
	Recursive bool

	// StopChan is an optional signal channel to stop listing early
	StopChan <-chan struct{}

	// Max keys
	MaxKeys int

	// Start token
	StartAfter string

	// Use ListObjectsV2
	UseV2 bool

	// Include object versions
	WithVersions bool

	// Include object metadata
	WithMetadata bool

	// Custom headers
	CustomHeaders http.Header
}

// CopyOptions controls server-side copy behavior
type CopyOptions struct {
	// Source version ID
	SourceVersionID string

	// Destination metadata and tags
	UserMetadata map[string]string
	UserTags     map[string]string

	// Replace metadata and/or tagging instead of copying existing
	ReplaceMetadata bool
	ReplaceTagging  bool

	// Object header overrides
	ContentType        string
	ContentEncoding    string
	ContentDisposition string
	CacheControl       string
	Expires            time.Time

	// Storage class
	StorageClass string

	// Conditional copy headers
	MatchETag     string
	NotMatchETag  string
	MatchModified time.Time
	NotModified   time.Time

	// Custom headers
	CustomHeaders http.Header
}

// WithContentType sets Content-Type
func WithContentType(contentType string) PutOption {
	return func(opts *PutOptions) {
		opts.ContentType = contentType
		opts.contentTypeSet = true
	}
}

// WithPresignHeaders adds headers that must be signed for the presigned URL
func WithPresignHeaders(headers http.Header) PresignOption {
	return func(opts *PresignOptions) {
		if opts.Headers == nil {
			opts.Headers = make(http.Header)
		}
		for k, v := range headers {
			opts.Headers[k] = append([]string{}, v...)
		}
	}
}

// WithPresignQuery adds additional query parameters (e.g., response-content-type)
func WithPresignQuery(values url.Values) PresignOption {
	return func(opts *PresignOptions) {
		if opts.QueryValues == nil {
			opts.QueryValues = make(url.Values)
		}
		for k, v := range values {
			for _, vv := range v {
				opts.QueryValues.Add(k, vv)
			}
		}
	}
}

// WithPresignSSES3 signs SSE-S3 header for presigned requests
func WithPresignSSES3() PresignOption {
	return func(opts *PresignOptions) {
		if opts.Headers == nil {
			opts.Headers = make(http.Header)
		}
		opts.Headers.Set("x-amz-server-side-encryption", "AES256")
	}
}

// WithPresignSSECustomer signs SSE-C headers for presigned requests (key must be base64 encoded)
func WithPresignSSECustomer(keyB64, keyMD5 string) PresignOption {
	return func(opts *PresignOptions) {
		if opts.Headers == nil {
			opts.Headers = make(http.Header)
		}
		opts.Headers.Set("x-amz-server-side-encryption-customer-algorithm", "AES256")
		opts.Headers.Set("x-amz-server-side-encryption-customer-key", keyB64)
		if keyMD5 != "" {
			opts.Headers.Set("x-amz-server-side-encryption-customer-key-MD5", keyMD5)
		}
	}
}

// WithContentEncoding sets Content-Encoding
func WithContentEncoding(encoding string) PutOption {
	return func(opts *PutOptions) {
		opts.ContentEncoding = encoding
	}
}

// WithContentDisposition sets Content-Disposition
func WithContentDisposition(disposition string) PutOption {
	return func(opts *PutOptions) {
		opts.ContentDisposition = disposition
	}
}

// WithUserMetadata sets user metadata
func WithUserMetadata(metadata map[string]string) PutOption {
	return func(opts *PutOptions) {
		opts.UserMetadata = metadata
	}
}

// WithUserTags sets object tags
func WithUserTags(tags map[string]string) PutOption {
	return func(opts *PutOptions) {
		opts.UserTags = tags
	}
}

// WithStorageClass sets storage class
func WithStorageClass(class string) PutOption {
	return func(opts *PutOptions) {
		opts.StorageClass = class
	}
}

// WithPartSize sets multipart part size
func WithPartSize(size uint64) PutOption {
	return func(opts *PutOptions) {
		opts.PartSize = size
	}
}

// WithSSES3 enables SSE-S3 server-side encryption for uploads
func WithSSES3() PutOption {
	return func(opts *PutOptions) {
		if opts.CustomHeaders == nil {
			opts.CustomHeaders = make(http.Header)
		}
		opts.CustomHeaders.Set("x-amz-server-side-encryption", "AES256")
	}
}

// WithSSECustomer provides SSE-C parameters for uploads (key must be base64 encoded)
func WithSSECustomer(keyB64, keyMD5 string) PutOption {
	return func(opts *PutOptions) {
		opts.SSECustomerAlgorithm = "AES256"
		opts.SSECustomerKey = keyB64
		opts.SSECustomerKeyMD5 = keyMD5
	}
}

// WithGetRange sets byte range for downloads
func WithGetRange(start, end int64) GetOption {
	return func(opts *GetOptions) {
		opts.RangeStart = start
		opts.RangeEnd = end
		opts.SetRange = true
	}
}

// WithGetSSECustomer sets SSE-C parameters for downloads (key must be base64 encoded)
func WithGetSSECustomer(keyB64, keyMD5 string) GetOption {
	return func(opts *GetOptions) {
		opts.SSECustomerAlgorithm = "AES256"
		opts.SSECustomerKey = keyB64
		opts.SSECustomerKeyMD5 = keyMD5
	}
}

// WithVersionID selects a specific object version (Get/Stat/Delete)
func WithVersionID(versionID string) interface{} {
	return struct {
		GetOption
		StatOption
		DeleteOption
	}{
		GetOption: func(opts *GetOptions) {
			opts.VersionID = versionID
		},
		StatOption: func(opts *StatOptions) {
			opts.VersionID = versionID
		},
		DeleteOption: func(opts *DeleteOptions) {
			opts.VersionID = versionID
		},
	}
}

// WithListPrefix sets listing prefix
func WithListPrefix(prefix string) ListOption {
	return func(opts *ListOptions) {
		opts.Prefix = prefix
	}
}

// WithListRecursive toggles recursive listing
func WithListRecursive(recursive bool) ListOption {
	return func(opts *ListOptions) {
		opts.Recursive = recursive
	}
}

// WithListStopChan sets a channel to stop listing early
func WithListStopChan(ch <-chan struct{}) ListOption {
	return func(opts *ListOptions) {
		opts.StopChan = ch
	}
}

// WithListMaxKeys sets the maximum keys to return
func WithListMaxKeys(maxKeys int) ListOption {
	return func(opts *ListOptions) {
		opts.MaxKeys = maxKeys
	}
}

// WithCopySourceVersionID sets the source version ID for a copy
func WithCopySourceVersionID(versionID string) CopyOption {
	return func(opts *CopyOptions) {
		opts.SourceVersionID = versionID
	}
}

// WithCopyMetadata sets destination metadata for copy
func WithCopyMetadata(metadata map[string]string, replace bool) CopyOption {
	return func(opts *CopyOptions) {
		opts.UserMetadata = metadata
		opts.ReplaceMetadata = replace
	}
}
