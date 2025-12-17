// Package signer internal/signer/sts.go
// STS specific signing functions
package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

const (
	serviceTypeSTS = "sts"
)

// SignV4STS 为 STS 请求签名（用于 AssumeRole 等操作）
// 这个函数独立于其他 signer 包的功能，避免循环依赖
func SignV4STS(req http.Request, accessKeyID, secretAccessKey, location string) *http.Request {
	// 匿名凭证不需要签名
	if accessKeyID == "" || secretAccessKey == "" {
		return &req
	}

	// 设置时间
	t := time.Now().UTC()
	req.Header.Set("X-Amz-Date", t.Format(iso8601DateFormat))

	// 确保有 Host 头
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", getHostAddr(&req))
	}

	// 对于 STS，使用特定的 region
	region := location
	if region == "" {
		region = "us-east-1"
	}

	// 使用 serviceTypeSTS 构建 scope
	scope := buildCredentialScopeForSTS(t, region)
	canonicalRequest := buildCanonicalRequestForSTS(&req)
	stringToSign := buildStringToSignForSTS(canonicalRequest, t, scope)
	signingKey := deriveSigningKeyForSTS(secretAccessKey, t, region)
	signature := hex.EncodeToString(hmacSHA256ForSTS(signingKey, []byte(stringToSign)))

	// 构建 Authorization 头
	signedHeaders := getSignedHeadersForSTS(req.Header)
	authorization := signV4Algorithm + " " +
		"Credential=" + accessKeyID + "/" + scope + ", " +
		"SignedHeaders=" + signedHeaders + ", " +
		"Signature=" + signature

	req.Header.Set("Authorization", authorization)

	return &req
}

// buildCredentialScopeForSTS 构建 STS 的凭证范围
func buildCredentialScopeForSTS(t time.Time, region string) string {
	return strings.Join([]string{
		t.Format(yyyymmdd),
		region,
		serviceTypeSTS,
		"aws4_request",
	}, "/")
}

// deriveSigningKeyForSTS 派生 STS 签名密钥
func deriveSigningKeyForSTS(secretAccessKey string, t time.Time, region string) []byte {
	kSecret := []byte("AWS4" + secretAccessKey)
	kDate := hmacSHA256ForSTS(kSecret, []byte(t.Format(yyyymmdd)))
	kRegion := hmacSHA256ForSTS(kDate, []byte(region))
	kService := hmacSHA256ForSTS(kRegion, []byte(serviceTypeSTS))
	kSigning := hmacSHA256ForSTS(kService, []byte("aws4_request"))
	return kSigning
}

// hmacSHA256ForSTS 计算 HMAC-SHA256（独立函数，避免依赖 v4.go）
func hmacSHA256ForSTS(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// buildCanonicalRequestForSTS 构建 STS 的标准请求
func buildCanonicalRequestForSTS(req *http.Request) string {
	// 获取签名的头部列表
	signedHeaders := getSignedHeadersForSTS(req.Header)

	// 构建标准头部字符串
	var canonicalHeaders strings.Builder
	headers := strings.Split(signedHeaders, ";")
	for _, h := range headers {
		canonicalHeaders.WriteString(h)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(strings.TrimSpace(req.Header.Get(h)))
		canonicalHeaders.WriteString("\n")
	}

	// 获取编码后的路径
	encodedPath := req.URL.EscapedPath()
	if encodedPath == "" {
		encodedPath = "/"
	}

	// 获取查询字符串
	canonicalQuery := req.URL.Query().Encode()

	// 组合标准请求
	return strings.Join([]string{
		req.Method,
		encodedPath,
		canonicalQuery,
		canonicalHeaders.String(),
		signedHeaders,
		req.Header.Get("X-Amz-Content-Sha256"),
	}, "\n")
}

// buildStringToSignForSTS 构建 STS 的待签名字符串
func buildStringToSignForSTS(canonicalRequest string, t time.Time, scope string) string {
	hash := sha256.Sum256([]byte(canonicalRequest))
	return strings.Join([]string{
		signV4Algorithm,
		t.Format(iso8601DateFormat),
		scope,
		hex.EncodeToString(hash[:]),
	}, "\n")
}

// getSignedHeadersForSTS 获取需要签名的头部列表
func getSignedHeadersForSTS(header http.Header) string {
	var headers []string
	for k := range header {
		lowerKey := strings.ToLower(k)
		// 跳过某些不需要签名的头部
		if lowerKey == "authorization" || lowerKey == "user-agent" {
			continue
		}
		headers = append(headers, lowerKey)
	}

	// 对头部排序
	sort := func(s []string) {
		for i := 0; i < len(s); i++ {
			for j := i + 1; j < len(s); j++ {
				if s[i] > s[j] {
					s[i], s[j] = s[j], s[i]
				}
			}
		}
	}
	sort(headers)

	return strings.Join(headers, ";")
}
