// Package types/object.go
package types

import (
	"net/http"
	"time"
)

// ObjectInfo contains object metadata information
type ObjectInfo struct {
	// Basic information
	Key          string    `json:"name"`
	Size         int64     `json:"size"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"contentType"`
	LastModified time.Time `json:"lastModified"`
	Expires      time.Time `json:"expires,omitempty"`

	// Owner
	Owner Owner `json:"owner,omitempty"`

	// Storage class
	StorageClass string `json:"storageClass,omitempty"`

	// Version information
	VersionID      string `json:"versionId,omitempty"`
	IsLatest       bool   `json:"isLatest,omitempty"`
	IsDeleteMarker bool   `json:"isDeleteMarker,omitempty"`

	// Replication status
	ReplicationStatus string `json:"replicationStatus,omitempty"`

	// Metadata
	Metadata     http.Header `json:"metadata,omitempty"`
	UserMetadata StringMap   `json:"userMetadata,omitempty"`
	UserTags     URLMap      `json:"userTags,omitempty"`
	UserTagCount int         `json:"userTagCount,omitempty"`

	// Lifecycle
	Expiration       time.Time `json:"expiration,omitempty"`
	ExpirationRuleID string    `json:"expirationRuleId,omitempty"`

	// Restore information
	Restore *RestoreInfo `json:"restore,omitempty"`

	// Checksums
	ChecksumCRC32     string `json:"checksumCRC32,omitempty"`
	ChecksumCRC32C    string `json:"checksumCRC32C,omitempty"`
	ChecksumSHA1      string `json:"checksumSHA1,omitempty"`
	ChecksumSHA256    string `json:"checksumSHA256,omitempty"`
	ChecksumCRC64NVME string `json:"checksumCRC64NVME,omitempty"`
	ChecksumMode      string `json:"checksumMode,omitempty"`

	// ACL
	Grant []Grant `json:"grant,omitempty"`

	// Number of versions
	NumVersions int `json:"numVersions,omitempty"`

	// Internal information (EC encoding)
	Internal *struct {
		K int
		M int
	} `json:"internal,omitempty"`

	// Error (used for list operations)
	Err error `json:"-"`

	// IsPrefix indicates this entry is a common prefix (pseudo-directory)
	IsPrefix bool `json:"isPrefix,omitempty"`
}

// ObjectToDelete represents an object to be deleted
type ObjectToDelete struct {
	Key       string
	VersionID string
}

// DeletedObject contains deleted object result
type DeletedObject struct {
	Key                   string
	VersionID             string
	DeleteMarker          bool
	DeleteMarkerVersionID string
}

// DeleteError represents a delete error
type DeleteError struct {
	Key       string
	VersionID string
	Code      string
	Message   string
}

// CopyInfo contains copy information
type CopyInfo struct {
	Bucket            string
	Key               string
	ETag              string
	VersionID         string
	SourceVersionID   string
	LastModified      time.Time
	ChecksumCRC32     string
	ChecksumCRC32C    string
	ChecksumSHA1      string
	ChecksumSHA256    string
	ChecksumCRC64NVME string
}
