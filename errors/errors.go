// Package errors/errors.go
package errors

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// Error RustFSGo error interface
type Error interface {
	error
	Code() RustfsGoErrorCode
	Message() string
	StatusCode() int
	RequestID() string
	Resource() string
}

// APIError S3 API error
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

// Code returns the error code
func (e *APIError) Code() RustfsGoErrorCode {
	return RustfsGoErrorCode(e.ErrorCode)
}

// Message returns the error message
func (e *APIError) Message() string {
	return e.ErrorMessage
}

// StatusCode returns the HTTP status code
func (e *APIError) StatusCode() int {
	return e.ErrorStatusCode
}

// RequestID returns the request ID
func (e *APIError) RequestID() string {
	return e.ErrorRequestID
}

// Resource returns the resource
func (e *APIError) Resource() string {
	return e.ErrorResource
}

// NewAPIError creates a new APIError
func NewAPIError(code RustfsGoErrorCode, message string, statusCode int) *APIError {
	return &APIError{
		ErrorCode:       code.Error(),
		ErrorMessage:    message,
		ErrorStatusCode: statusCode,
	}
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.ErrorRequestID != "" {
		return fmt.Sprintf("%s: %s (RequestID: %s)", e.ErrorCode, e.ErrorMessage, e.ErrorRequestID)
	}
	return fmt.Sprintf("%s: %s", e.ErrorCode, e.ErrorMessage)
}

// WithRequestID sets the request ID
func (e *APIError) WithRequestID(id string) *APIError {
	e.ErrorRequestID = id
	return e
}

// WithResource sets the resource
func (e *APIError) WithResource(resource string) *APIError {
	e.ErrorResource = resource
	return e
}

// WithRegion sets the region
func (e *APIError) WithRegion(region string) *APIError {
	e.Region = region
	return e
}

// ParseErrorResponse parses an error response from the server
func ParseErrorResponse(resp *http.Response, bucketName, objectName string) error {
	if resp == nil {
		return NewAPIError(ErrCodeInternalError, "empty response", 500)
	}

	// read response body
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // Maxsize 1MB
	if err != nil {
		return NewAPIError(ErrCodeInternalError, "failed to read response body", resp.StatusCode)
	}

	// try to unmarshal XML error
	apiErr := &APIError{ErrorStatusCode: resp.StatusCode}
	if len(body) > 0 {
		if xmlErr := xml.Unmarshal(body, apiErr); xmlErr == nil {
			apiErr.ErrorStatusCode = resp.StatusCode
			return apiErr
		}
	}

	// create generic error
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
