// Package errors/codes.go
package errors

// RustfsGoErrorCode defines RustfsGo error codes
type RustfsGoErrorCode string

// S3 standard error codes
const (
	// bucket
	ErrCodeNoSuchBucket            RustfsGoErrorCode = "NoSuchBucket"
	ErrCodeBucketAlreadyExists     RustfsGoErrorCode = "BucketAlreadyExists"
	ErrCodeBucketAlreadyOwnedByYou RustfsGoErrorCode = "BucketAlreadyOwnedByYou"
	ErrCodeBucketNotEmpty          RustfsGoErrorCode = "BucketNotEmpty"
	ErrCodeInvalidBucketName       RustfsGoErrorCode = "InvalidBucketName"

	// object
	ErrCodeNoSuchKey         RustfsGoErrorCode = "NoSuchKey"
	ErrCodeInvalidObjectName RustfsGoErrorCode = "XRustfsInvalidObjectName"
	ErrCodeNoSuchUpload      RustfsGoErrorCode = "NoSuchUpload"
	ErrCodeNoSuchVersion     RustfsGoErrorCode = "NoSuchVersion"
	ErrCodeInvalidPart       RustfsGoErrorCode = "InvalidPart"
	ErrCodeInvalidPartOrder  RustfsGoErrorCode = "InvalidPartOrder"
	ErrCodeEntityTooLarge    RustfsGoErrorCode = "EntityTooLarge"
	ErrCodeEntityTooSmall    RustfsGoErrorCode = "EntityTooSmall"

	// access
	ErrCodeAccessDenied          RustfsGoErrorCode = "AccessDenied"
	ErrCodeAccountProblem        RustfsGoErrorCode = "AccountProblem"
	ErrCodeInvalidAccessKeyId    RustfsGoErrorCode = "InvalidAccessKeyId"
	ErrCodeSignatureDoesNotMatch RustfsGoErrorCode = "SignatureDoesNotMatch"

	// request
	ErrCodeInvalidArgument      RustfsGoErrorCode = "InvalidArgument"
	ErrCodeInvalidRequest       RustfsGoErrorCode = "InvalidRequest"
	ErrCodeMalformedXML         RustfsGoErrorCode = "MalformedXML"
	ErrCodeMissingContentLength RustfsGoErrorCode = "MissingContentLength"
	ErrCodeMethodNotAllowed     RustfsGoErrorCode = "MethodNotAllowed"
	ErrNilResponse              RustfsGoErrorCode = "NilResponse"
	ErrRequestTimeout           RustfsGoErrorCode = "RequestTimeout"
	ErrRequestTimeTooSkewed     RustfsGoErrorCode = "RequestTimeTooSkewed"
	ErrMovedPermanently         RustfsGoErrorCode = "MovedPermanently"
	ErrConflict                 RustfsGoErrorCode = "Conflict"
	ErrInvalidRange             RustfsGoErrorCode = "InvalidRange"

	// region and authorization
	ErrCodeInvalidRegion                RustfsGoErrorCode = "InvalidRegion"
	ErrCodeAuthorizationHeaderMalformed RustfsGoErrorCode = "AuthorizationHeaderMalformed"

	// server
	ErrCodeInternalError      RustfsGoErrorCode = "InternalError"
	ErrCodeServiceUnavailable RustfsGoErrorCode = "ServiceUnavailable"
	ErrCodeSlowDown           RustfsGoErrorCode = "SlowDown"
	ErrCodeNotImplemented     RustfsGoErrorCode = "NotImplemented"

	// preconditions
	ErrCodePreconditionFailed RustfsGoErrorCode = "PreconditionFailed"
	ErrCodeNotModified        RustfsGoErrorCode = "NotModified"

	// copy
	ErrCodeInvalidCopySource RustfsGoErrorCode = "InvalidCopySource"
)

// Error implements the error interface for RustfsGoErrorCode
func (e RustfsGoErrorCode) Error() string {
	return string(e)
}

// HTTP status code to RustfsGoErrorCode mapping
var httpStatusToCode = map[int]RustfsGoErrorCode{
	301: ErrMovedPermanently,
	400: ErrCodeInvalidArgument,
	403: ErrCodeAccessDenied,
	404: ErrCodeNoSuchKey,
	405: ErrCodeMethodNotAllowed,
	409: ErrConflict,
	411: ErrCodeMissingContentLength,
	412: ErrCodePreconditionFailed,
	416: ErrInvalidRange,
	500: ErrCodeInternalError,
	501: ErrCodeNotImplemented,
	503: ErrCodeServiceUnavailable,
}
