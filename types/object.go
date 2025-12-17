// Package types/object.go
package types

import (
	"net/http"
	"time"
)

// ObjectInfo 对象元数据信息
type ObjectInfo struct {
	// 基本信息
	Key          string    `json:"name"`
	Size         int64     `json:"size"`
	ETag         string    `json:"etag"`
	ContentType  string    `json:"contentType"`
	LastModified time.Time `json:"lastModified"`
	Expires      time.Time `json:"expires,omitempty"`

	// 所有者
	Owner Owner `json:"owner,omitempty"`

	// 存储类
	StorageClass string `json:"storageClass,omitempty"`

	// 版本信息
	VersionID      string `json:"versionId,omitempty"`
	IsLatest       bool   `json:"isLatest,omitempty"`
	IsDeleteMarker bool   `json:"isDeleteMarker,omitempty"`

	// 复制状态
	ReplicationStatus string `json:"replicationStatus,omitempty"`

	// 元数据
	Metadata     http.Header `json:"metadata,omitempty"`
	UserMetadata StringMap   `json:"userMetadata,omitempty"`
	UserTags     URLMap      `json:"userTags,omitempty"`
	UserTagCount int         `json:"userTagCount,omitempty"`

	// 生命周期
	Expiration       time.Time `json:"expiration,omitempty"`
	ExpirationRuleID string    `json:"expirationRuleId,omitempty"`

	// 恢复信息
	Restore *RestoreInfo `json:"restore,omitempty"`

	// 校验和
	ChecksumCRC32     string `json:"checksumCRC32,omitempty"`
	ChecksumCRC32C    string `json:"checksumCRC32C,omitempty"`
	ChecksumSHA1      string `json:"checksumSHA1,omitempty"`
	ChecksumSHA256    string `json:"checksumSHA256,omitempty"`
	ChecksumCRC64NVME string `json:"checksumCRC64NVME,omitempty"`
	ChecksumMode      string `json:"checksumMode,omitempty"`

	// ACL
	Grant []Grant `json:"grant,omitempty"`

	// 版本数量
	NumVersions int `json:"numVersions,omitempty"`

	// 内部信息（EC 编码）
	Internal *struct {
		K int
		M int
	} `json:"internal,omitempty"`

	// 错误（用于列表操作）
	Err error `json:"-"`
}

// ObjectToDelete 待删除对象
type ObjectToDelete struct {
	Key       string
	VersionID string
}

// DeletedObject 已删除对象结果
type DeletedObject struct {
	Key                   string
	VersionID             string
	DeleteMarker          bool
	DeleteMarkerVersionID string
}

// DeleteError 删除错误
type DeleteError struct {
	Key       string
	VersionID string
	Code      string
	Message   string
}

// CopyInfo 复制信息
type CopyInfo struct {
	Bucket          string
	Key             string
	ETag            string
	VersionID       string
	SourceVersionID string
	LastModified    time.Time
	ChecksumCRC32   string
	ChecksumCRC32C  string
	ChecksumSHA1    string
	ChecksumSHA256  string
	ChecksumCRC64NVME string
}