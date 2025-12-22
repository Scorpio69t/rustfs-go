// Package object object/utils.go
package object

import (
	"io"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/errors"
)

// closeResponse closes the HTTP response body
func closeResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}

// parseErrorResponse parses error response
func parseErrorResponse(resp *http.Response, bucketName, objectName string) error {
	return errors.ParseErrorResponse(resp, bucketName, objectName)
}
