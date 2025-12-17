// Package types/upload.go
package types

import "time"

// UploadInfo 上传结果信息
type UploadInfo struct {
	// 桶名称
	Bucket string `json:"bucket"`
	// 对象键
	Key string `json:"key"`
	// ETag
	ETag string `json:"etag"`
	// 大小
	Size int64 `json:"size"`
	// 最后修改时间
	LastModified time.Time `json:"lastModified"`
	// 位置
	Location string `json:"location,omitempty"`
	// 版本 ID
	VersionID string `json:"versionId,omitempty"`

	// 生命周期过期信息
	Expiration       time.Time `json:"expiration,omitempty"`
	ExpirationRuleID string    `json:"expirationRuleId,omitempty"`

	// 校验和
	ChecksumCRC32     string `json:"checksumCRC32,omitempty"`
	ChecksumCRC32C    string `json:"checksumCRC32C,omitempty"`
	ChecksumSHA1      string `json:"checksumSHA1,omitempty"`
	ChecksumSHA256    string `json:"checksumSHA256,omitempty"`
	ChecksumCRC64NVME string `json:"checksumCRC64NVME,omitempty"`
	ChecksumMode      string `json:"checksumMode,omitempty"`
}

// MultipartInfo 分片上传信息
type MultipartInfo struct {
	// 上传 ID
	UploadID string `json:"uploadId"`
	// 对象键
	Key string `json:"key"`
	// 发起时间
	Initiated time.Time `json:"initiated"`
	// 发起者
	Initiator struct {
		ID          string
		DisplayName string
	} `json:"initiator,omitempty"`
	// 所有者
	Owner Owner `json:"owner,omitempty"`
	// 存储类
	StorageClass string `json:"storageClass,omitempty"`
	// 大小（聚合）
	Size int64 `json:"size,omitempty"`
	// 错误
	Err error `json:"-"`
}

// PartInfo 分片信息
type PartInfo struct {
	// 分片号
	PartNumber int `json:"partNumber"`
	// ETag
	ETag string `json:"etag"`
	// 大小
	Size int64 `json:"size"`
	// 最后修改时间
	LastModified time.Time `json:"lastModified"`

	// 校验和
	ChecksumCRC32     string `json:"checksumCRC32,omitempty"`
	ChecksumCRC32C    string `json:"checksumCRC32C,omitempty"`
	ChecksumSHA1      string `json:"checksumSHA1,omitempty"`
	ChecksumSHA256    string `json:"checksumSHA256,omitempty"`
	ChecksumCRC64NVME string `json:"checksumCRC64NVME,omitempty"`
}

// CompletePart 完成分片信息
type CompletePart struct {
	PartNumber        int
	ETag              string
	ChecksumCRC32     string
	ChecksumCRC32C    string
	ChecksumSHA1      string
	ChecksumSHA256    string
	ChecksumCRC64NVME string
}
