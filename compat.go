// Package rustfs compat.go - 向后兼容层
package rustfs

import (
	"github.com/Scorpio69t/rustfs-go/object"
)

// --- 兼容性类型定义 ---

// MakeBucketOptions 创建存储桶选项（兼容旧 API）
type MakeBucketOptions struct {
	// Region 区域
	Region string
	// ObjectLocking 启用对象锁定
	ObjectLocking bool
}

// RemoveBucketOptions 删除存储桶选项（兼容旧 API）
type RemoveBucketOptions struct {
	// ForceDelete 强制删除
	ForceDelete bool
}

// PutObjectOptions 上传对象选项（兼容旧 API）
type PutObjectOptions struct {
	ContentType        string
	ContentEncoding    string
	ContentDisposition string
	ContentLanguage    string
	CacheControl       string
	UserMetadata       map[string]string
	UserTags           map[string]string
	StorageClass       string
	PartSize           uint64
	NumThreads         uint
	SendContentMD5     bool
}

// GetObjectOptions 下载对象选项（兼容旧 API）
type GetObjectOptions struct {
	VersionID string
}

// StatObjectOptions 获取对象信息选项（兼容旧 API）
type StatObjectOptions struct {
	VersionID string
}

// RemoveObjectOptions 删除对象选项（兼容旧 API）
type RemoveObjectOptions struct {
	VersionID   string
	ForceDelete bool
}

// ListObjectsOptions 列出对象选项（兼容旧 API）
type ListObjectsOptions struct {
	Prefix     string
	Recursive  bool
	StartAfter string
	MaxKeys    int
}

// CopyObjectOptions 复制对象选项（兼容旧 API）
type CopyObjectOptions struct {
	SourceVersionID string
	ReplaceMetadata bool
	ContentType     string
	UserMetadata    map[string]string
	StorageClass    string
}

// --- 选项转换函数 ---

// convertPutOptions 转换上传选项
func convertPutOptions(opts PutObjectOptions) []object.PutOption {
	var options []object.PutOption

	if opts.ContentType != "" {
		options = append(options, object.WithContentType(opts.ContentType))
	}
	if opts.ContentEncoding != "" {
		options = append(options, object.WithContentEncoding(opts.ContentEncoding))
	}
	if opts.ContentDisposition != "" {
		options = append(options, object.WithContentDisposition(opts.ContentDisposition))
	}
	// ContentLanguage 和 CacheControl 暂时不支持，可以通过 CustomHeaders 添加
	// TODO: 添加 WithContentLanguage 和 WithCacheControl 选项函数
	if opts.UserMetadata != nil {
		options = append(options, object.WithUserMetadata(opts.UserMetadata))
	}
	if opts.UserTags != nil {
		options = append(options, object.WithUserTags(opts.UserTags))
	}
	if opts.StorageClass != "" {
		options = append(options, object.WithStorageClass(opts.StorageClass))
	}
	if opts.PartSize > 0 {
		options = append(options, object.WithPartSize(opts.PartSize))
	}

	return options
}

// convertGetOptions 转换下载选项
func convertGetOptions(opts GetObjectOptions) []object.GetOption {
	var options []object.GetOption

	if opts.VersionID != "" {
		// WithVersionID 返回一个复合类型，需要类型断言
		versionOpt := object.WithVersionID(opts.VersionID)
		if getOpt, ok := versionOpt.(object.GetOption); ok {
			options = append(options, getOpt)
		}
	}

	return options
}

// convertStatOptions 转换获取对象信息选项
func convertStatOptions(opts StatObjectOptions) []object.StatOption {
	var options []object.StatOption

	if opts.VersionID != "" {
		// WithVersionID 返回一个复合类型，需要类型断言
		versionOpt := object.WithVersionID(opts.VersionID)
		if statOpt, ok := versionOpt.(object.StatOption); ok {
			options = append(options, statOpt)
		}
	}

	return options
}

// convertDeleteOptions 转换删除选项
func convertDeleteOptions(opts RemoveObjectOptions) []object.DeleteOption {
	var options []object.DeleteOption

	if opts.VersionID != "" {
		// WithVersionID 返回一个复合类型，需要类型断言
		versionOpt := object.WithVersionID(opts.VersionID)
		if delOpt, ok := versionOpt.(object.DeleteOption); ok {
			options = append(options, delOpt)
		}
	}

	return options
}

// convertListOptions 转换列出对象选项
func convertListOptions(opts ListObjectsOptions) []object.ListOption {
	var options []object.ListOption

	if opts.Prefix != "" {
		options = append(options, object.WithListPrefix(opts.Prefix))
	}
	if opts.Recursive {
		options = append(options, object.WithListRecursive(opts.Recursive))
	}
	if opts.MaxKeys > 0 {
		options = append(options, object.WithListMaxKeys(opts.MaxKeys))
	}

	return options
}

// convertCopyOptions 转换复制选项
func convertCopyOptions(opts CopyObjectOptions) []object.CopyOption {
	var options []object.CopyOption

	if opts.SourceVersionID != "" {
		options = append(options, object.WithCopySourceVersionID(opts.SourceVersionID))
	}
	if opts.ReplaceMetadata {
		options = append(options, object.WithCopyMetadata(opts.UserMetadata, opts.ReplaceMetadata))
	}

	return options
}
