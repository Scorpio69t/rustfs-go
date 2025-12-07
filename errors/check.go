// Package errors/check.go
package errors

import "errors"

// IsNotFound 检查是否为未找到错误
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code() == ErrCodeNoSuchBucket ||
			apiErr.Code() == ErrCodeNoSuchKey ||
			apiErr.Code() == ErrCodeNoSuchUpload
	}
	return false
}

// IsBucketNotFound 检查桶是否不存在
func IsBucketNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code() == ErrCodeNoSuchBucket
	}
	return false
}

// IsObjectNotFound 检查对象是否不存在
func IsObjectNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code() == ErrCodeNoSuchKey
	}
	return false
}

// IsAccessDenied 检查是否为访问拒绝错误
func IsAccessDenied(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code() == ErrCodeAccessDenied
	}
	return false
}

// IsBucketExists 检查桶是否已存在
func IsBucketExists(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code() == ErrCodeBucketAlreadyExists ||
			apiErr.Code() == ErrCodeBucketAlreadyOwnedByYou
	}
	return false
}

// IsBucketNotEmpty 检查桶是否非空
func IsBucketNotEmpty(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code() == ErrCodeBucketNotEmpty
	}
	return false
}

// IsInvalidArgument 检查是否为无效参数错误
func IsInvalidArgument(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code() == ErrCodeInvalidArgument
	}
	return false
}

// IsServiceUnavailable 检查服务是否不可用
func IsServiceUnavailable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code() == ErrCodeServiceUnavailable ||
			apiErr.Code() == ErrCodeSlowDown
	}
	return false
}

// IsRetryable 检查错误是否可重试
func IsRetryable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch apiErr.Code() {
		case ErrCodeServiceUnavailable,
			ErrCodeSlowDown,
			ErrCodeInternalError,
			"RequestTimeout",
			"RequestTimeTooSkewed":
			return true
		}
		// 5xx 错误通常可重试
		if apiErr.StatusCode() >= 500 {
			return true
		}
	}
	return false
}

// ToAPIError 将错误转换为 APIError
func ToAPIError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}
