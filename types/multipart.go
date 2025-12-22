// Package types types/multipart.go - Multipart upload related types
package types

// ObjectPart contains object part information (used for multipart upload)
type ObjectPart struct {
	// PartNumber is the part number (starts from 1)
	PartNumber int `xml:"PartNumber"`
	// ETag is the part's ETag
	ETag string `xml:"ETag"`
	// Size is the part size
	Size int64 `xml:"Size,omitempty"`
	// LastModified is the last modified time
	LastModified string `xml:"LastModified,omitempty"`

	// Checksums
	ChecksumCRC32  string `xml:"ChecksumCRC32,omitempty"`
	ChecksumCRC32C string `xml:"ChecksumCRC32C,omitempty"`
	ChecksumSHA1   string `xml:"ChecksumSHA1,omitempty"`
	ChecksumSHA256 string `xml:"ChecksumSHA256,omitempty"`
}
