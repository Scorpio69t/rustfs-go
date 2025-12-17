// Package signer internal/signer/v4.go
package signer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Signature and API related constants
const (
	signV4Algorithm   = "AWS4-HMAC-SHA256"
	iso8601DateFormat = "20060102T150405Z"
	yyyymmdd          = "20060102"
	serviceTypeS3     = "s3"
)

// v4IgnoredHeaders 签名时忽略的头部
// 参考: https://github.com/aws/aws-sdk-js/issues/659#issuecomment-120477258
var v4IgnoredHeaders = map[string]bool{
	"Accept-Encoding": true,
	"Authorization":   true,
	"User-Agent":      true,
}

// V4Signer AWS Signature Version 4 签名器
type V4Signer struct{}

// Sign 使用 V4 算法签名请求
func (s *V4Signer) Sign(req *http.Request, accessKey, secretKey, sessionToken, region string) *http.Request {
	// Presign 不需要签名
	if accessKey == "" || secretKey == "" {
		return req
	}

	// 设置时间
	t := time.Now().UTC()
	req.Header.Set("X-Amz-Date", t.Format(iso8601DateFormat))

	// 设置 session token
	if sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", sessionToken)
	}

	// 确保有 Host 头
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", getHostAddr(req))
	}

	// 计算签名
	signature := s.calculateSignature(req, accessKey, secretKey, region, t)

	// 构建 Authorization 头
	auth := s.buildAuthorizationHeader(req, accessKey, region, signature, t)
	req.Header.Set("Authorization", auth)

	return req
}

// Presign 使用 V4 算法预签名请求
// 参考: http://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-query-string-auth.html
func (s *V4Signer) Presign(req *http.Request, accessKey, secretKey, sessionToken, region string, expires time.Duration) *http.Request {
	// Presign 不需要签名
	if accessKey == "" || secretKey == "" {
		return req
	}

	// 初始化时间
	t := time.Now().UTC()

	// 获取凭证字符串
	credential := s.getCredential(accessKey, region, t)

	// 获取所有签名头
	signedHeaders := s.getSignedHeaders(req.Header)

	// 设置 URL 查询参数
	query := req.URL.Query()
	query.Set("X-Amz-Algorithm", signV4Algorithm)
	query.Set("X-Amz-Date", t.Format(iso8601DateFormat))
	query.Set("X-Amz-Expires", strconv.FormatInt(int64(expires.Seconds()), 10))
	query.Set("X-Amz-SignedHeaders", signedHeaders)
	query.Set("X-Amz-Credential", credential)

	// 设置 session token（如果有）
	if sessionToken != "" {
		query.Set("X-Amz-Security-Token", sessionToken)
	}

	req.URL.RawQuery = query.Encode()

	// 获取规范请求
	canonicalRequest := s.createCanonicalRequest(req)

	// 获取待签名字符串
	stringToSign := s.createStringToSign(canonicalRequest, region, t)

	// 获取签名密钥
	signingKey := s.deriveSigningKey(secretKey, region, t)

	// 计算签名
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	// 将签名添加到查询参数
	req.URL.RawQuery += "&X-Amz-Signature=" + signature

	return req
}

// getCredential 生成凭证字符串
func (s *V4Signer) getCredential(accessKeyID, region string, t time.Time) string {
	scope := s.credentialScope(region, t)
	return accessKeyID + "/" + scope
}

// calculateSignature 计算签名
func (s *V4Signer) calculateSignature(req *http.Request, accessKey, secretKey, region string, t time.Time) string {
	// 1. 创建规范请求
	canonicalRequest := s.createCanonicalRequest(req)

	// 2. 创建待签名字符串
	stringToSign := s.createStringToSign(canonicalRequest, region, t)

	// 3. 计算签名
	signingKey := s.deriveSigningKey(secretKey, region, t)
	signature := hmacSHA256(signingKey, []byte(stringToSign))

	return hex.EncodeToString(signature)
}

// createCanonicalRequest 创建规范请求
// 格式: <HTTPMethod>\n<CanonicalURI>\n<CanonicalQueryString>\n<CanonicalHeaders>\n<SignedHeaders>\n<HashedPayload>
func (s *V4Signer) createCanonicalRequest(req *http.Request) string {
	// HTTP Method
	method := req.Method

	// Canonical URI - URL 编码路径
	uri := encodePath(req.URL.Path)
	if uri == "" {
		uri = "/"
	}

	// Canonical Query String - 替换 + 为 %20
	req.URL.RawQuery = strings.ReplaceAll(req.URL.Query().Encode(), "+", "%20")

	// Canonical Headers
	canonicalHeaders := s.getCanonicalHeaders(req)

	// Signed Headers
	signedHeaders := s.getSignedHeaders(req.Header)

	// Payload Hash
	payloadHash := s.getHashedPayload(req)

	return strings.Join([]string{
		method,
		uri,
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
}

// getHashedPayload 获取请求负载的哈希值
func (s *V4Signer) getHashedPayload(req *http.Request) string {
	hashedPayload := req.Header.Get("X-Amz-Content-Sha256")
	if hashedPayload == "" {
		// Presign 没有 payload，使用 S3 推荐值
		hashedPayload = UnsignedPayload
	}
	return hashedPayload
}

// getCanonicalHeaders 生成规范头部列表
func (s *V4Signer) getCanonicalHeaders(req *http.Request) string {
	var headers []string
	vals := make(map[string][]string)

	for k, vv := range req.Header {
		if _, ok := v4IgnoredHeaders[http.CanonicalHeaderKey(k)]; ok {
			continue // 忽略的头部
		}
		lowerKey := strings.ToLower(k)
		headers = append(headers, lowerKey)
		vals[lowerKey] = vv
	}

	// 确保包含 host 头
	if !headerExists("host", headers) {
		headers = append(headers, "host")
	}
	sort.Strings(headers)

	var buf bytes.Buffer
	// 保存所有头部为规范格式 <header>:<value> 每个头部换行分隔
	for _, k := range headers {
		buf.WriteString(k)
		buf.WriteByte(':')
		switch k {
		case "host":
			buf.WriteString(getHostAddr(req))
			buf.WriteByte('\n')
		default:
			for idx, v := range vals[k] {
				if idx > 0 {
					buf.WriteByte(',')
				}
				buf.WriteString(signV4TrimAll(v))
			}
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}

// getSignedHeaders 生成所有签名请求头
// 返回按字典序排序、分号分隔的小写请求头名称列表
func (s *V4Signer) getSignedHeaders(header http.Header) string {
	var headers []string
	for k := range header {
		if _, ok := v4IgnoredHeaders[http.CanonicalHeaderKey(k)]; ok {
			continue // 忽略的头部
		}
		headers = append(headers, strings.ToLower(k))
	}
	if !headerExists("host", headers) {
		headers = append(headers, "host")
	}
	sort.Strings(headers)
	return strings.Join(headers, ";")
}

// createStringToSign 创建待签名字符串
func (s *V4Signer) createStringToSign(canonicalRequest, region string, t time.Time) string {
	scope := s.credentialScope(region, t)
	hash := sha256.Sum256([]byte(canonicalRequest))
	return strings.Join([]string{
		"AWS4-HMAC-SHA256",
		t.Format("20060102T150405Z"),
		scope,
		hex.EncodeToString(hash[:]),
	}, "\n")
}

// credentialScope 创建凭证范围
func (s *V4Signer) credentialScope(region string, t time.Time) string {
	return strings.Join([]string{
		t.Format("20060102"),
		region,
		"s3",
		"aws4_request",
	}, "/")
}

// deriveSigningKey 派生签名密钥
func (s *V4Signer) deriveSigningKey(secretKey, region string, t time.Time) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secretKey), []byte(t.Format("20060102")))
	regionKey := hmacSHA256(dateKey, []byte(region))
	serviceKey := hmacSHA256(regionKey, []byte("s3"))
	signingKey := hmacSHA256(serviceKey, []byte("aws4_request"))
	return signingKey
}

// buildAuthorizationHeader 构建 Authorization 头
func (s *V4Signer) buildAuthorizationHeader(req *http.Request, accessKey, region, signature string, t time.Time) string {
	signedHeaders := s.getSignedHeaders(req.Header)
	scope := s.credentialScope(region, t)

	return signV4Algorithm + " " +
		"Credential=" + accessKey + "/" + scope + ", " +
		"SignedHeaders=" + signedHeaders + ", " +
		"Signature=" + signature
}

// hmacSHA256 计算 HMAC-SHA256
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
