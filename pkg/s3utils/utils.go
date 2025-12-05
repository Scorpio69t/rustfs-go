// Package s3utils provides utility functions for S3 operations
package s3utils

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const (
	// MinPartSize - minimum part size per multipart
	MinPartSize = 1024 * 1024 * 5 // 5MB

	// MaxPartSize - maximum part size per multipart
	MaxPartSize = 1024 * 1024 * 1024 * 5 // 5GB

	// MaxObjectSize - maximum object size
	MaxObjectSize = 1024 * 1024 * 1024 * 1024 * 5 // 5TB
)

var (
	// ValidBucketNameRegex - valid bucket name regex
	ValidBucketNameRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]$`)

	// ValidObjectNameRegex - valid object name regex
	ValidObjectNameRegex = regexp.MustCompile(`^[a-zA-Z0-9!_.*'()-\/]+$`)
)

// CheckValidBucketName - checks if bucket name is valid
func CheckValidBucketName(bucketName string) error {
	if strings.TrimSpace(bucketName) == "" {
		return errors.New("bucket name cannot be empty")
	}
	if len(bucketName) < 3 {
		return errors.New("bucket name cannot be shorter than 3 characters")
	}
	if len(bucketName) > 63 {
		return errors.New("bucket name cannot be longer than 63 characters")
	}
	if !ValidBucketNameRegex.MatchString(bucketName) {
		return errors.New("bucket name contains invalid characters")
	}
	return nil
}

// CheckValidObjectName - checks if object name is valid
func CheckValidObjectName(objectName string) error {
	if strings.TrimSpace(objectName) == "" {
		return errors.New("object name cannot be empty")
	}
	if len(objectName) > 1024 {
		return errors.New("object name cannot be longer than 1024 characters")
	}
	if !ValidObjectNameRegex.MatchString(objectName) {
		return errors.New("object name contains invalid characters")
	}
	return nil
}

// EncodePath - encode path
func EncodePath(pathName string) string {
	if pathName == "" {
		return ""
	}
	encodedPath := "/"
	for _, v := range strings.Split(pathName, "/") {
		encodedPath += url.PathEscape(v) + "/"
	}
	// Remove trailing '/'
	return strings.TrimSuffix(encodedPath, "/")
}

// GetMD5Hash - get MD5 hash of data
func GetMD5Hash(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// IsValidRegion - check if region is valid
func IsValidRegion(region string) bool {
	if region == "" {
		return false
	}
	// Basic validation - can be extended
	return len(region) >= 2 && len(region) <= 50
}

// URLEncodePath - URL encode path
func URLEncodePath(pathName string) string {
	if pathName == "" {
		return ""
	}
	return url.PathEscape(pathName)
}

// URLDecodePath - URL decode path
func URLDecodePath(encodedPath string) (string, error) {
	return url.PathUnescape(encodedPath)
}

// ValidatePartSize - validate part size
func ValidatePartSize(size int64) error {
	if size < MinPartSize {
		return fmt.Errorf("part size must be at least %d bytes", MinPartSize)
	}
	if size > MaxPartSize {
		return fmt.Errorf("part size must be at most %d bytes", MaxPartSize)
	}
	return nil
}
