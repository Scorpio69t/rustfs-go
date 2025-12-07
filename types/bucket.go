// Package types/bucket.go
package types

import "time"

// BucketInfo 桶信息
type BucketInfo struct {
	// 桶名称
	Name string `json:"name"`
	// 创建时间
	CreationDate time.Time `json:"creationDate"`
	// 桶所在区域
	Region string `json:"region,omitempty"`
}

// BucketLookupType 桶查找类型
type BucketLookupType int

const (
	// BucketLookupAuto 自动检测
	BucketLookupAuto BucketLookupType = iota
	// BucketLookupDNS DNS 风格
	BucketLookupDNS
	// BucketLookupPath 路径风格
	BucketLookupPath
)

// VersioningConfig 版本控制配置
type VersioningConfig struct {
	Status    string // Enabled, Suspended
	MFADelete string // Enabled, Disabled
}

// IsEnabled 检查版本控制是否启用
func (v VersioningConfig) IsEnabled() bool {
	return v.Status == "Enabled"
}

// IsSuspended 检查版本控制是否暂停
func (v VersioningConfig) IsSuspended() bool {
	return v.Status == "Suspended"
}
