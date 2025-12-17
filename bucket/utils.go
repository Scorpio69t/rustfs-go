// Package bucket bucket/utils.go
package bucket

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/Scorpio69t/rustfs-go/errors"
)

// sumSHA256Hex 计算 SHA256 哈希并返回十六进制字符串
func sumSHA256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// closeResponse 关闭响应
func closeResponse(resp *http.Response) {
	if resp != nil && resp.Body != nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}

// parseErrorResponse 解析错误响应
func parseErrorResponse(resp *http.Response, bucketName, objectName string) error {
	return errors.ParseErrorResponse(resp, bucketName, objectName)
}
