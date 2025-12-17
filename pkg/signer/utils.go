// Package signer internal/signer/utils.go
package signer

import (
	"crypto/sha256"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// Constants for unsigned payload
const (
	UnsignedPayload = "UNSIGNED-PAYLOAD"
)

// sum256 计算 SHA256 哈希
func sum256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// getHostAddr 返回 host header，如果不存在则返回 URL 中的 host
func getHostAddr(req *http.Request) string {
	host := req.Header.Get("Host")
	if host != "" && req.Host != host {
		return host
	}
	if req.Host != "" {
		return req.Host
	}
	return req.URL.Host
}

// signV4TrimAll 压缩连续空格为一个空格（按照 AWS Signature V4 规范）
// http://docs.aws.amazon.com/general/latest/gr/sigv4-create-canonical-request.html
func signV4TrimAll(input string) string {
	// 使用 strings.Fields 会自动 trim 并压缩空格
	return strings.Join(strings.Fields(input), " ")
}

// encodePath URL 编码路径（保留 /）
func encodePath(pathName string) string {
	if pathName == "" {
		return "/"
	}

	// 保留尾部斜杠
	trailingSlash := strings.HasSuffix(pathName, "/")

	// S3 要求保留路径中的斜杠，但编码其他特殊字符
	var encodedPathname strings.Builder
	for _, s := range strings.Split(pathName, "/") {
		if len(s) == 0 {
			continue
		}
		encodedPathname.WriteString("/")
		encodedPathname.WriteString(url.PathEscape(s))
	}

	path := encodedPathname.String()
	if len(path) == 0 {
		path = "/"
	}

	// 如果原路径有尾部斜杠且不是根路径，保留它
	if trailingSlash && path != "/" {
		path += "/"
	}

	return path
}

// queryEncode 编码查询参数（用于预签名 URL）
func queryEncode(query url.Values) string {
	// 对查询参数按键排序
	keys := make([]string, 0, len(query))
	for k := range query {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	for _, k := range keys {
		vs := query[k]
		keyEscaped := url.QueryEscape(k)
		for _, v := range vs {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(keyEscaped)
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(v))
		}
	}
	return buf.String()
}

// headerExists 检查 header 是否存在
func headerExists(key string, headers []string) bool {
	for _, k := range headers {
		if k == key {
			return true
		}
	}
	return false
}
