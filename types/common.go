// Package types/common.go
package types

import "time"

// Owner contains object owner information
type Owner struct {
	DisplayName string `json:"displayName,omitempty"`
	ID          string `json:"id,omitempty"`
}

// Grant represents ACL grant
type Grant struct {
	Grantee    Grantee
	Permission string
}

// Grantee represents grantee information
type Grantee struct {
	Type        string
	ID          string
	DisplayName string
	URI         string
}

// RestoreInfo contains archive restore information
type RestoreInfo struct {
	OngoingRestore bool
	ExpiryTime     time.Time
}

// ChecksumType represents checksum type
type ChecksumType int

const (
	ChecksumNone ChecksumType = iota
	ChecksumCRC32
	ChecksumCRC32C
	ChecksumSHA1
	ChecksumSHA256
	ChecksumCRC64NVME
)

// String returns the checksum type as a string
func (c ChecksumType) String() string {
	switch c {
	case ChecksumCRC32:
		return "CRC32"
	case ChecksumCRC32C:
		return "CRC32C"
	case ChecksumSHA1:
		return "SHA1"
	case ChecksumSHA256:
		return "SHA256"
	case ChecksumCRC64NVME:
		return "CRC64NVME"
	default:
		return ""
	}
}

// RetentionMode represents retention mode
type RetentionMode string

const (
	RetentionGovernance RetentionMode = "GOVERNANCE"
	RetentionCompliance RetentionMode = "COMPLIANCE"
)

// IsValid checks if the retention mode is valid
func (r RetentionMode) IsValid() bool {
	return r == RetentionGovernance || r == RetentionCompliance
}

// LegalHoldStatus represents legal hold status
type LegalHoldStatus string

const (
	LegalHoldOn  LegalHoldStatus = "ON"
	LegalHoldOff LegalHoldStatus = "OFF"
)

// IsValid checks if the legal hold status is valid
func (l LegalHoldStatus) IsValid() bool {
	return l == LegalHoldOn || l == LegalHoldOff
}

// ReplicationStatus represents replication status
type ReplicationStatus string

const (
	ReplicationPending  ReplicationStatus = "PENDING"
	ReplicationComplete ReplicationStatus = "COMPLETED"
	ReplicationFailed   ReplicationStatus = "FAILED"
	ReplicationReplica  ReplicationStatus = "REPLICA"
)

// StringMap is a custom string map (used for XML parsing)
type StringMap map[string]string

// URLMap is a URL-encoded map
type URLMap map[string]string
