// Package object object/options.go
package object

import (
	"net/http"
	"time"
)

// PutOptions 上传选项
type PutOptions struct {
	// 内容类型
	ContentType string

	// 内容编码
	ContentEncoding string

	// 内容处置
	ContentDisposition string

	// 内容语言
	ContentLanguage string

	// 缓存控制
	CacheControl string

	// 过期时间
	Expires time.Time

	// 用户元数据
	UserMetadata map[string]string

	// 用户标签
	UserTags map[string]string

	// 存储类
	StorageClass string

	// 自定义头
	CustomHeaders http.Header

	// 发送 Content-MD5
	SendContentMD5 bool

	// 禁用 Content-SHA256
	DisableContentSHA256 bool

	// 分片大小
	PartSize uint64

	// 并发数
	NumThreads uint
}

// GetOptions 下载选项
type GetOptions struct {
	// Range 请求范围
	RangeStart int64
	RangeEnd   int64
	SetRange   bool

	// 版本 ID
	VersionID string

	// 匹配条件
	MatchETag     string
	NotMatchETag  string
	MatchModified time.Time
	NotModified   time.Time

	// 自定义头
	CustomHeaders http.Header
}

// StatOptions 获取对象信息选项
type StatOptions struct {
	// 版本 ID
	VersionID string

	// 自定义头
	CustomHeaders http.Header
}

// DeleteOptions 删除选项
type DeleteOptions struct {
	// 版本 ID
	VersionID string

	// 强制删除
	ForceDelete bool

	// 自定义头
	CustomHeaders http.Header
}

// ListOptions 列出对象选项
type ListOptions struct {
	// 前缀
	Prefix string

	// 递归列出
	Recursive bool

	// 最大键数
	MaxKeys int

	// 起始位置
	StartAfter string

	// 使用 V2 API
	UseV2 bool

	// 包含版本
	WithVersions bool

	// 包含元数据
	WithMetadata bool
}

// CopyOptions 复制选项
type CopyOptions struct {
	// 源版本 ID
	SourceVersionID string

	// 目标元数据
	UserMetadata map[string]string

	// 替换元数据
	ReplaceMetadata bool

	// 内容类型
	ContentType string

	// 存储类
	StorageClass string

	// 匹配条件
	MatchETag     string
	NotMatchETag  string
	MatchModified time.Time
	NotModified   time.Time

	// 自定义头
	CustomHeaders http.Header
}

// WithContentType 设置内容类型
func WithContentType(contentType string) PutOption {
	return func(opts *PutOptions) {
		opts.ContentType = contentType
	}
}

// WithContentEncoding 设置内容编码
func WithContentEncoding(encoding string) PutOption {
	return func(opts *PutOptions) {
		opts.ContentEncoding = encoding
	}
}

// WithContentDisposition 设置内容处置
func WithContentDisposition(disposition string) PutOption {
	return func(opts *PutOptions) {
		opts.ContentDisposition = disposition
	}
}

// WithUserMetadata 设置用户元数据
func WithUserMetadata(metadata map[string]string) PutOption {
	return func(opts *PutOptions) {
		opts.UserMetadata = metadata
	}
}

// WithUserTags 设置用户标签
func WithUserTags(tags map[string]string) PutOption {
	return func(opts *PutOptions) {
		opts.UserTags = tags
	}
}

// WithStorageClass 设置存储类
func WithStorageClass(class string) PutOption {
	return func(opts *PutOptions) {
		opts.StorageClass = class
	}
}

// WithPartSize 设置分片大小
func WithPartSize(size uint64) PutOption {
	return func(opts *PutOptions) {
		opts.PartSize = size
	}
}

// WithGetRange 设置下载范围
func WithGetRange(start, end int64) GetOption {
	return func(opts *GetOptions) {
		opts.RangeStart = start
		opts.RangeEnd = end
		opts.SetRange = true
	}
}

// WithVersionID 设置版本 ID（用于 Get/Stat/Delete）
func WithVersionID(versionID string) interface{} {
	return struct {
		GetOption
		StatOption
		DeleteOption
	}{
		GetOption: func(opts *GetOptions) {
			opts.VersionID = versionID
		},
		StatOption: func(opts *StatOptions) {
			opts.VersionID = versionID
		},
		DeleteOption: func(opts *DeleteOptions) {
			opts.VersionID = versionID
		},
	}
}

// WithListPrefix 设置列出前缀
func WithListPrefix(prefix string) ListOption {
	return func(opts *ListOptions) {
		opts.Prefix = prefix
	}
}

// WithListRecursive 设置递归列出
func WithListRecursive(recursive bool) ListOption {
	return func(opts *ListOptions) {
		opts.Recursive = recursive
	}
}

// WithListMaxKeys 设置最大键数
func WithListMaxKeys(maxKeys int) ListOption {
	return func(opts *ListOptions) {
		opts.MaxKeys = maxKeys
	}
}

// WithCopySourceVersionID 设置源版本 ID
func WithCopySourceVersionID(versionID string) CopyOption {
	return func(opts *CopyOptions) {
		opts.SourceVersionID = versionID
	}
}

// WithCopyMetadata 设置复制元数据
func WithCopyMetadata(metadata map[string]string, replace bool) CopyOption {
	return func(opts *CopyOptions) {
		opts.UserMetadata = metadata
		opts.ReplaceMetadata = replace
	}
}
