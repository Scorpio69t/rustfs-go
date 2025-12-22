// Package types/bucket.go
package types

import "time"

// BucketInfo contains bucket information
type BucketInfo struct {
	// Bucket name
	Name string `json:"name"`
	// Creation date
	CreationDate time.Time `json:"creationDate"`
	// Bucket region
	Region string `json:"region,omitempty"`
}

// BucketLookupType represents bucket lookup type
type BucketLookupType int

const (
	// BucketLookupAuto automatically detects
	BucketLookupAuto BucketLookupType = iota
	// BucketLookupDNS uses DNS style
	BucketLookupDNS
	// BucketLookupPath uses path style
	BucketLookupPath
)

// VersioningConfig contains versioning configuration
type VersioningConfig struct {
	Status    string // Enabled, Suspended
	MFADelete string // Enabled, Disabled
}

// IsEnabled checks if versioning is enabled
func (v VersioningConfig) IsEnabled() bool {
	return v.Status == "Enabled"
}

// IsSuspended checks if versioning is suspended
func (v VersioningConfig) IsSuspended() bool {
	return v.Status == "Suspended"
}
