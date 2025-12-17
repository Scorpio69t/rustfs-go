// Package types/common.go
package types

import "time"

// Owner 对象所有者信息
type Owner struct {
	DisplayName string `json:"displayName,omitempty"`
	ID          string `json:"id,omitempty"`
}

// Grant ACL 授权
type Grant struct {
	Grantee    Grantee
	Permission string
}

// Grantee 授权对象
type Grantee struct {
	Type        string
	ID          string
	DisplayName string
	URI         string
}

// RestoreInfo 归档恢复信息
type RestoreInfo struct {
	OngoingRestore bool
	ExpiryTime     time.Time
}

// ChecksumType 校验和类型
type ChecksumType int

const (
	ChecksumNone ChecksumType = iota
	ChecksumCRC32
	ChecksumCRC32C
	ChecksumSHA1
	ChecksumSHA256
	ChecksumCRC64NVME
)

// String 返回校验和类型字符串
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

// RetentionMode 保留模式
type RetentionMode string

const (
	RetentionGovernance RetentionMode = "GOVERNANCE"
	RetentionCompliance RetentionMode = "COMPLIANCE"
)

// IsValid 验证保留模式是否有效
func (r RetentionMode) IsValid() bool {
	return r == RetentionGovernance || r == RetentionCompliance
}

// LegalHoldStatus 法律保留状态
type LegalHoldStatus string

const (
	LegalHoldOn  LegalHoldStatus = "ON"
	LegalHoldOff LegalHoldStatus = "OFF"
)

// IsValid 验证法律保留状态是否有效
func (l LegalHoldStatus) IsValid() bool {
	return l == LegalHoldOn || l == LegalHoldOff
}

// ReplicationStatus 复制状态
type ReplicationStatus string

const (
	ReplicationPending  ReplicationStatus = "PENDING"
	ReplicationComplete ReplicationStatus = "COMPLETED"
	ReplicationFailed   ReplicationStatus = "FAILED"
	ReplicationReplica  ReplicationStatus = "REPLICA"
)

// StringMap 自定义字符串映射（用于 XML 解析）
type StringMap map[string]string

// URLMap URL 编码的映射
type URLMap map[string]string
