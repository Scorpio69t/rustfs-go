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
func (s *objectService) FPut(ctx context.Context, bucketName, objectName, filePath string, opts ...PutOption) (info types.UploadInfo, err error) {
	if err := validateBucketName(bucketName); err != nil {
		return info, err
	}
	if err := validateObjectName(objectName); err != nil {
		return info, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return info, err
	}
	defer func() {
		if cerr := file.Close(); err == nil && cerr != nil {
			err = cerr
		}
	}()

	stat, err := file.Stat()
	if err != nil {
		return info, err
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
	// Re-apply options to include inferred content type
	_ = applyPutOptions(opts)

	reader := io.NewSectionReader(file, 0, stat.Size())
	info, err = s.Put(ctx, bucketName, objectName, reader, stat.Size(), opts...)
	return info, err
}

// FGet downloads an object directly to a local file path.
func (s *objectService) FGet(ctx context.Context, bucketName, objectName, filePath string, opts ...GetOption) (info types.ObjectInfo, err error) {
	if err := validateBucketName(bucketName); err != nil {
		return info, err
	}
	if err := validateObjectName(objectName); err != nil {
		return info, err
	}

	var reader io.ReadCloser
	reader, info, err = s.Get(ctx, bucketName, objectName, opts...)
	if err != nil {
		return info, err
	}
	defer func() {
		if cerr := reader.Close(); err == nil && cerr != nil {
			err = cerr
		}
	}()

	dir := filepath.Dir(filePath)
	if dir == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return info, err
	}

	tmpFile, err := os.CreateTemp(dir, ".fget-*")
	if err != nil {
		return info, err
	}

	success := false
	defer func() {
		if !success {
			if cerr := tmpFile.Close(); err == nil && cerr != nil {
				err = cerr
			}
			if rerr := os.Remove(tmpFile.Name()); err == nil && rerr != nil && !os.IsNotExist(rerr) {
				err = rerr
			}
		}
	}()

	if _, err = io.Copy(tmpFile, reader); err != nil {
		return info, err
	}

	if err := tmpFile.Close(); err != nil {
		return info, err
	}

	if err := os.Rename(tmpFile.Name(), filePath); err != nil {
		return info, err
	}

	success = true
	return info, err
}
