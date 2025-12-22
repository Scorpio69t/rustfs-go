// Package bucket bucket/utils.go
package bucket

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/errors"
)

// sumSHA256Hex computes the SHA256 hash of the given data and returns it as a hexadecimal string.
func sumSHA256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// closeResponse closes the HTTP response body safely.
func closeResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}

// parseErrorResponse parses the error response from an HTTP response and returns a structured error.
func parseErrorResponse(resp *http.Response, bucketName, objectName string) error {
	return errors.ParseErrorResponse(resp, bucketName, objectName)
}
