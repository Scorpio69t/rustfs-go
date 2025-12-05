// Package s3errors provides S3 error handling
package s3errors

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

// ErrorResponse - error response format
type ErrorResponse struct {
	XMLName    xml.Name `xml:"Error"`
	Code       string   `xml:"Code"`
	Message    string   `xml:"Message"`
	BucketName string   `xml:"BucketName"`
	Key        string   `xml:"Key"`
	Resource   string   `xml:"Resource"`
	RequestID  string   `xml:"RequestId"`
	HostID     string   `xml:"HostId"`
}

// Error - returns error string
func (e ErrorResponse) Error() string {
	return e.Message
}

// ToErrorResponse - Returns parsed error response
func ToErrorResponse(resp *http.Response, bucketName, objectName string) error {
	if resp == nil {
		return fmt.Errorf("response is nil")
	}

	errorResponse := ErrorResponse{
		BucketName: bucketName,
		Key:        objectName,
		Resource:   fmt.Sprintf("/%s/%s", bucketName, objectName),
	}

	err := xml.NewDecoder(resp.Body).Decode(&errorResponse)
	if err != nil && err != io.EOF {
		return fmt.Errorf("parse error response: %w", err)
	}

	if errorResponse.Code == "" {
		errorResponse.Code = resp.Status
		errorResponse.Message = http.StatusText(resp.StatusCode)
	}

	return errorResponse
}
