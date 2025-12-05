package rustfs

import (
	"context"
	"mime"
	"os"
	"path/filepath"

	"github.com/Scorpio69t/rustfs-go/pkg/s3utils"
)

// FPutObject - Create an object in a bucket, with contents from file at filePath. Allows request cancellation.
func (c *Client) FPutObject(ctx context.Context, bucketName, objectName, filePath string, opts PutObjectOptions) (info UploadInfo, err error) {
	// Input validation.
	if err := s3utils.CheckValidBucketName(bucketName); err != nil {
		return UploadInfo{}, err
	}
	if err := s3utils.CheckValidObjectName(objectName); err != nil {
		return UploadInfo{}, err
	}

	// Open the referenced file.
	fileReader, err := os.Open(filePath)
	// If any error fail quickly here.
	if err != nil {
		return UploadInfo{}, err
	}
	defer fileReader.Close()

	// Save the file stat.
	fileStat, err := fileReader.Stat()
	if err != nil {
		return UploadInfo{}, err
	}

	// Save the file size.
	fileSize := fileStat.Size()

	// Set contentType based on filepath extension if not given or default
	// value of "application/octet-stream" if the extension has no associated type.
	if opts.ContentType == "" {
		if opts.ContentType = mime.TypeByExtension(filepath.Ext(filePath)); opts.ContentType == "" {
			opts.ContentType = "application/octet-stream"
		}
	}
	return c.PutObject(ctx, bucketName, objectName, fileReader, fileSize, opts)
}
