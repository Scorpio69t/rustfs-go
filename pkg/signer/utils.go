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

// sum256 computes SHA256 hash
func sum256(data []byte) []byte {
	hash := sha256.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// getHostAddr returns host header, or URL host if missing
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

// signV4TrimAll collapses consecutive spaces to one (per AWS SigV4 spec)
// http://docs.aws.amazon.com/general/latest/gr/sigv4-create-canonical-request.html
func signV4TrimAll(input string) string {
	// strings.Fields auto trims and collapses spaces
	return strings.Join(strings.Fields(input), " ")
}

// encodePath URL-encodes path (preserves /)
func encodePath(pathName string) string {
	if pathName == "" {
		return "/"
	}

	// Preserve trailing slash
	trailingSlash := strings.HasSuffix(pathName, "/")

	// S3 requires keeping slashes while encoding other special chars
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

	// Keep trailing slash if original path had it and not root
	if trailingSlash && path != "/" {
		path += "/"
	}

	return path
}

// queryEncode encodes query params (for presigned URLs)
func queryEncode(query url.Values) string {
	// Sort query parameters by key
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

// headerExists checks if header exists
func headerExists(key string, headers []string) bool {
	for _, k := range headers {
		if k == key {
			return true
		}
	}
	return false
}
