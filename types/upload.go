// Package types/upload.go
package types

import "time"

// UploadInfo contains upload result information
type UploadInfo struct {
	// Bucket name
	Bucket string `json:"bucket"`
	// Object key
	Key string `json:"key"`
	// ETag
	ETag string `json:"etag"`
	// Size
	Size int64 `json:"size"`
	// Last modified time
	LastModified time.Time `json:"lastModified"`
	// Location
	Location string `json:"location,omitempty"`
	// Version ID
	VersionID string `json:"versionId,omitempty"`

	// Lifecycle expiration information
	Expiration       time.Time `json:"expiration,omitempty"`
	ExpirationRuleID string    `json:"expirationRuleId,omitempty"`

	// Checksums
	ChecksumCRC32     string `json:"checksumCRC32,omitempty"`
	ChecksumCRC32C    string `json:"checksumCRC32C,omitempty"`
	ChecksumSHA1      string `json:"checksumSHA1,omitempty"`
	ChecksumSHA256    string `json:"checksumSHA256,omitempty"`
	ChecksumCRC64NVME string `json:"checksumCRC64NVME,omitempty"`
	ChecksumMode      string `json:"checksumMode,omitempty"`
}

// MultipartInfo contains multipart upload information
type MultipartInfo struct {
	// Upload ID
	UploadID string `json:"uploadId"`
	// Object key
	Key string `json:"key"`
	// Initiated time
	Initiated time.Time `json:"initiated"`
	// Initiator
	Initiator struct {
		ID          string
		DisplayName string
	} `json:"initiator,omitempty"`
	// Owner
	Owner Owner `json:"owner,omitempty"`
	// Storage class
	StorageClass string `json:"storageClass,omitempty"`
	// Size (aggregated)
	Size int64 `json:"size,omitempty"`
	// Error
	Err error `json:"-"`
}

// PartInfo contains part information
type PartInfo struct {
	// Part number
	PartNumber int `json:"partNumber"`
	// ETag
	ETag string `json:"etag"`
	// Size
	Size int64 `json:"size"`
	// Last modified time
	LastModified time.Time `json:"lastModified"`

	// Checksums
	ChecksumCRC32     string `json:"checksumCRC32,omitempty"`
	ChecksumCRC32C    string `json:"checksumCRC32C,omitempty"`
	ChecksumSHA1      string `json:"checksumSHA1,omitempty"`
	ChecksumSHA256    string `json:"checksumSHA256,omitempty"`
	ChecksumCRC64NVME string `json:"checksumCRC64NVME,omitempty"`
}

// CompletePart contains completed part information
type CompletePart struct {
	PartNumber        int
	ETag              string
	ChecksumCRC32     string
	ChecksumCRC32C    string
	ChecksumSHA1      string
	ChecksumSHA256    string
	ChecksumCRC64NVME string
}
