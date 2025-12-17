// Package errors/errors.go
package errors

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// Error RustFS 错误接口
type Error interface {
	error
	Code() RustfsGoErrorCode
	Message() string
	StatusCode() int
	RequestID() string
	Resource() string
}

// APIError S3 API 错误
type APIError struct {
	XMLName         xml.Name `xml:"Error"`
	ErrorCode       string   `xml:"Code"`
	ErrorMessage    string   `xml:"Message"`
	ErrorResource   string   `xml:"Resource"`
	ErrorRequestID  string   `xml:"RequestId"`
	HostID          string   `xml:"HostId"`
	ErrorStatusCode int      `xml:"-"`
	Region          string   `xml:"Region"`
}

// Code 返回错误码
func (e *APIError) Code() RustfsGoErrorCode {
	return RustfsGoErrorCode(e.ErrorCode)
}

// Message 返回错误消息
func (e *APIError) Message() string {
	return e.ErrorMessage
}

// StatusCode 返回 HTTP 状态码
func (e *APIError) StatusCode() int {
	return e.ErrorStatusCode
}

// RequestID 返回请求 ID
func (e *APIError) RequestID() string {
	return e.ErrorRequestID
}

// Resource 返回资源
func (e *APIError) Resource() string {
	return e.ErrorResource
}

// NewAPIError 创建新的 API 错误
func NewAPIError(code RustfsGoErrorCode, message string, statusCode int) *APIError {
	return &APIError{
		ErrorCode:       code.Error(),
		ErrorMessage:    message,
		ErrorStatusCode: statusCode,
	}
}

// Error 实现 error 接口
func (e *APIError) Error() string {
	if e.ErrorRequestID != "" {
		return fmt.Sprintf("%s: %s (RequestID: %s)", e.ErrorCode, e.ErrorMessage, e.ErrorRequestID)
	}
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.ErrorMessage)
}

// WithRequestID 设置请求 ID
func (e *APIError) WithRequestID(id string) *APIError {
	e.ErrorRequestID = id
	return e
}

// WithResource 设置资源
func (e *APIError) WithResource(resource string) *APIError {
	e.ErrorResource = resource
	return e
}

// WithRegion 设置区域
func (e *APIError) WithRegion(region string) *APIError {
	e.Region = region
	return e
}

// ParseErrorResponse 从 HTTP 响应解析错误
func ParseErrorResponse(resp *http.Response, bucketName, objectName string) error {
	if resp == nil {
		return NewAPIError(ErrCodeInternalError, "empty response", 500)
	}

	// 读取响应体
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 最大 1MB
	if err != nil {
		return NewAPIError(ErrCodeInternalError, "failed to read response body", resp.StatusCode)
	}

	// 尝试解析 XML 错误响应
	apiErr := &APIError{ErrorStatusCode: resp.StatusCode}
	if len(body) > 0 {
		if xmlErr := xml.Unmarshal(body, apiErr); xmlErr == nil {
			apiErr.ErrorStatusCode = resp.StatusCode
			return apiErr
		}
	}

	// 使用状态码生成错误
	code := httpStatusToCode[resp.StatusCode]
	if code == "" {
		code = ErrCodeInternalError
	}

	return &APIError{
		ErrorCode:       code.Error(),
		ErrorMessage:    http.StatusText(resp.StatusCode),
		ErrorStatusCode: resp.StatusCode,
		ErrorRequestID:  resp.Header.Get("x-amz-request-id"),
		HostID:          resp.Header.Get("x-amz-id-2"),
		ErrorResource:   "/" + bucketName + "/" + objectName,
	}
}
