// Package types types/multipart.go - 分片上传相关类型
package types

// ObjectPart 对象分片信息（用于分片上传）
type ObjectPart struct {
	// PartNumber 分片编号（从 1 开始）
	PartNumber int `xml:"PartNumber"`
	// ETag 分片的 ETag
	ETag string `xml:"ETag"`
	// Size 分片大小
	Size int64 `xml:"Size,omitempty"`
	// LastModified 最后修改时间
	LastModified string `xml:"LastModified,omitempty"`

	// Checksums 校验和
	ChecksumCRC32  string `xml:"ChecksumCRC32,omitempty"`
	ChecksumCRC32C string `xml:"ChecksumCRC32C,omitempty"`
	ChecksumSHA1   string `xml:"ChecksumSHA1,omitempty"`
	ChecksumSHA256 string `xml:"ChecksumSHA256,omitempty"`
}
