// Package errors/check.go
package errors

import (
	"errors"
)

// IsNotFound checks if the error indicates that a resource was not found
func IsNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return errors.Is(apiErr.Code(), ErrCodeNoSuchBucket) ||
			errors.Is(apiErr.Code(), ErrCodeNoSuchKey) ||
			errors.Is(apiErr.Code(), ErrCodeNoSuchUpload)
	}
	return false
}

// IsBucketNotFound checks if the bucket does not exist
func IsBucketNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return errors.Is(apiErr.Code(), ErrCodeNoSuchBucket)
	}
	return false
}

// IsObjectNotFound checks if the object does not exist
func IsObjectNotFound(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return errors.Is(apiErr.Code(), ErrCodeNoSuchKey)
	}
	return false
}

// IsAccessDenied checks if access is denied
func IsAccessDenied(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return errors.Is(apiErr.Code(), ErrCodeAccessDenied)
	}
	return false
}

// IsBucketExists checks if the bucket already exists
func IsBucketExists(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return errors.Is(apiErr.Code(), ErrCodeBucketAlreadyExists) ||
			errors.Is(apiErr.Code(), ErrCodeBucketAlreadyOwnedByYou)
	}
	return false
}

// IsBucketNotEmpty checks if the bucket is not empty
func IsBucketNotEmpty(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return errors.Is(apiErr.Code(), ErrCodeBucketNotEmpty)
	}
	return false
}

// IsInvalidArgument checks if the error is due to an invalid argument
func IsInvalidArgument(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return errors.Is(apiErr.Code(), ErrCodeInvalidArgument)
	}
	return false
}

// IsServiceUnavailable checks if the service is unavailable
func IsServiceUnavailable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return errors.Is(apiErr.Code(), ErrCodeServiceUnavailable) ||
			errors.Is(apiErr.Code(), ErrCodeSlowDown)
	}
	return false
}

// IsRetryable checks if the error is retryable
func IsRetryable(err error) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		switch err := apiErr.Code(); {
		case errors.Is(err, ErrCodeServiceUnavailable), errors.Is(err, ErrCodeSlowDown),
			errors.Is(err, ErrCodeInternalError), errors.Is(err, ErrRequestTimeout),
			errors.Is(err, ErrRequestTimeTooSkewed):
			return true
		}
		// 5xx errors usually indicate server-side issues
		if apiErr.StatusCode() >= 500 {
			return true
		}
	}
	return false
}

// ToAPIError converts a generic error to an APIError if possible
func ToAPIError(err error) *APIError {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}
