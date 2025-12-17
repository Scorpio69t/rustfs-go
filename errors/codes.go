// Package errors/codes.go
package errors

// RustfsGoErrorCode 定义了 RustfsGo 错误码的类型
type RustfsGoErrorCode string

// S3 标准错误码
const (
	// 桶相关
	ErrCodeNoSuchBucket            RustfsGoErrorCode = "NoSuchBucket"
	ErrCodeBucketAlreadyExists     RustfsGoErrorCode = "BucketAlreadyExists"
	ErrCodeBucketAlreadyOwnedByYou RustfsGoErrorCode = "BucketAlreadyOwnedByYou"
	ErrCodeBucketNotEmpty          RustfsGoErrorCode = "BucketNotEmpty"
	ErrCodeInvalidBucketName       RustfsGoErrorCode = "InvalidBucketName"

	// 对象相关
	ErrCodeNoSuchKey         RustfsGoErrorCode = "NoSuchKey"
	ErrCodeInvalidObjectName RustfsGoErrorCode = "XRustfsInvalidObjectName"
	ErrCodeNoSuchUpload      RustfsGoErrorCode = "NoSuchUpload"
	ErrCodeNoSuchVersion     RustfsGoErrorCode = "NoSuchVersion"
	ErrCodeInvalidPart       RustfsGoErrorCode = "InvalidPart"
	ErrCodeInvalidPartOrder  RustfsGoErrorCode = "InvalidPartOrder"
	ErrCodeEntityTooLarge    RustfsGoErrorCode = "EntityTooLarge"
	ErrCodeEntityTooSmall    RustfsGoErrorCode = "EntityTooSmall"

	// 访问控制
	ErrCodeAccessDenied          RustfsGoErrorCode = "AccessDenied"
	ErrCodeAccountProblem        RustfsGoErrorCode = "AccountProblem"
	ErrCodeInvalidAccessKeyId    RustfsGoErrorCode = "InvalidAccessKeyId"
	ErrCodeSignatureDoesNotMatch RustfsGoErrorCode = "SignatureDoesNotMatch"

	// 请求相关
	ErrCodeInvalidArgument      RustfsGoErrorCode = "InvalidArgument"
	ErrCodeInvalidRequest       RustfsGoErrorCode = "InvalidRequest"
	ErrCodeMalformedXML         RustfsGoErrorCode = "MalformedXML"
	ErrCodeMissingContentLength RustfsGoErrorCode = "MissingContentLength"
	ErrCodeMethodNotAllowed     RustfsGoErrorCode = "MethodNotAllowed"

	// 区域相关
	ErrCodeInvalidRegion                RustfsGoErrorCode = "InvalidRegion"
	ErrCodeAuthorizationHeaderMalformed RustfsGoErrorCode = "AuthorizationHeaderMalformed"

	// 服务器
	ErrCodeInternalError      RustfsGoErrorCode = "InternalError"
	ErrCodeServiceUnavailable RustfsGoErrorCode = "ServiceUnavailable"
	ErrCodeSlowDown           RustfsGoErrorCode = "SlowDown"
	ErrCodeNotImplemented     RustfsGoErrorCode = "NotImplemented"

	// 条件请求
	ErrCodePreconditionFailed RustfsGoErrorCode = "PreconditionFailed"
	ErrCodeNotModified        RustfsGoErrorCode = "NotModified"

	// 复制
	ErrCodeInvalidCopySource RustfsGoErrorCode = "InvalidCopySource"
)

// Error 实现 error 接口
func (e RustfsGoErrorCode) Error() string {
	return string(e)
}

// HTTP 状态码到错误码的映射
var httpStatusToCode = map[int]RustfsGoErrorCode{
	301: "MovedPermanently",
	400: ErrCodeInvalidArgument,
	403: ErrCodeAccessDenied,
	404: ErrCodeNoSuchKey,
	405: ErrCodeMethodNotAllowed,
	409: "Conflict",
	411: ErrCodeMissingContentLength,
	412: ErrCodePreconditionFailed,
	416: "InvalidRange",
	500: ErrCodeInternalError,
	501: ErrCodeNotImplemented,
	503: ErrCodeServiceUnavailable,
}
