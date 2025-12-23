// Package object object/fputget.go
package object

import (
	"context"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/Scorpio69t/rustfs-go/types"
)

// FPut uploads a local file using streaming IO and optional metadata.
func (s *objectService) FPut(ctx context.Context, bucketName, objectName, filePath string, opts ...PutOption) (types.UploadInfo, error) {
	if err := validateBucketName(bucketName); err != nil {
		return types.UploadInfo{}, err
	}
	if err := validateObjectName(objectName); err != nil {
		return types.UploadInfo{}, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return types.UploadInfo{}, err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return types.UploadInfo{}, err
	}

	// Detect content type from extension if caller did not set it
	options := applyPutOptions(opts)
	if !options.contentTypeSet {
		if ext := strings.ToLower(filepath.Ext(filePath)); ext != "" {
			if ct := mime.TypeByExtension(ext); ct != "" {
				opts = append(opts, WithContentType(ct))
			}
		}
	}

	return s.Put(ctx, bucketName, objectName, file, stat.Size(), opts...)
}

// FGet downloads an object directly to a local file path.
func (s *objectService) FGet(ctx context.Context, bucketName, objectName, filePath string, opts ...GetOption) (types.ObjectInfo, error) {
	if err := validateBucketName(bucketName); err != nil {
		return types.ObjectInfo{}, err
	}
	if err := validateObjectName(objectName); err != nil {
		return types.ObjectInfo{}, err
	}

	reader, info, err := s.Get(ctx, bucketName, objectName, opts...)
	if err != nil {
		return types.ObjectInfo{}, err
	}
	defer reader.Close()

	dir := filepath.Dir(filePath)
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return types.ObjectInfo{}, err
	}

	tmpFile, err := os.CreateTemp(dir, ".fget-*")
	if err != nil {
		return types.ObjectInfo{}, err
	}

	success := false
	defer func() {
		if !success {
			tmpFile.Close()
			os.Remove(tmpFile.Name())
		}
	}()

	if _, err = io.Copy(tmpFile, reader); err != nil {
		return types.ObjectInfo{}, err
	}

	if err := tmpFile.Close(); err != nil {
		return types.ObjectInfo{}, err
	}

	if err := os.Rename(tmpFile.Name(), filePath); err != nil {
		return types.ObjectInfo{}, err
	}

	success = true
	return info, nil
}
