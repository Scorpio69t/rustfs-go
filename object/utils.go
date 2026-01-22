// Package object object/utils.go
package object

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/errors"
	"github.com/Scorpio69t/rustfs-go/internal/core"
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

// applySSECustomerHeaders adds SSE-C headers to the request when provided.
func applySSECustomerHeaders(meta *core.RequestMetadata, algorithm, key, keyMD5 string) {
	if algorithm != "" && key != "" {
		meta.CustomHeader.Set("x-amz-server-side-encryption-customer-algorithm", algorithm)
		meta.CustomHeader.Set("x-amz-server-side-encryption-customer-key", key)
		if keyMD5 != "" {
			meta.CustomHeader.Set("x-amz-server-side-encryption-customer-key-MD5", keyMD5)
		}
	}
}

func sumMD5Base64(data []byte) string {
	sum := md5.Sum(data)
	return base64.StdEncoding.EncodeToString(sum[:])
}

func sumSHA256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
