// Package signer internal/signer/v4.go
package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sort"
	"strings"
	"time"
)

// V4Signer AWS Signature Version 4 签名器
type V4Signer struct{}

// Sign 使用 V4 算法签名请求
func (s *V4Signer) Sign(req *http.Request, accessKey, secretKey, sessionToken, region string) *http.Request {
	// 设置时间
	t := time.Now().UTC()
	req.Header.Set("X-Amz-Date", t.Format("20060102T150405Z"))

	// 设置 session token
	if sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", sessionToken)
	}

	// 计算签名
	signature := s.calculateSignature(req, accessKey, secretKey, region, t)

	// 构建 Authorization 头
	auth := s.buildAuthorizationHeader(req, accessKey, region, signature, t)
	req.Header.Set("Authorization", auth)

	return req
}

// Presign 使用 V4 算法预签名请求
func (s *V4Signer) Presign(req *http.Request, accessKey, secretKey, sessionToken, region string, expires time.Duration) *http.Request {
	// TODO: 实现预签名逻辑
	return req
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
func (s *V4Signer) createCanonicalRequest(req *http.Request) string {
	// HTTP Method
	method := req.Method

	// Canonical URI
	uri := req.URL.Path
	if uri == "" {
		uri = "/"
	}

	// Canonical Query String
	queryString := req.URL.Query().Encode()

	// Canonical Headers
	headers, signedHeaders := s.canonicalHeaders(req.Header)

	// Payload Hash
	payloadHash := req.Header.Get("X-Amz-Content-Sha256")
	if payloadHash == "" {
		payloadHash = "UNSIGNED-PAYLOAD"
	}

	return strings.Join([]string{
		method,
		uri,
		queryString,
		headers,
		signedHeaders,
		payloadHash,
	}, "\n")
}

// canonicalHeaders 创建规范头部
func (s *V4Signer) canonicalHeaders(header http.Header) (canonical, signed string) {
	var keys []string
	for k := range header {
		keys = append(keys, strings.ToLower(k))
	}
	sort.Strings(keys)

	var headers []string
	var signedHeaders []string
	for _, k := range keys {
		if k == "host" || strings.HasPrefix(k, "x-amz-") || k == "content-type" {
			headers = append(headers, k+":"+strings.TrimSpace(header.Get(k)))
			signedHeaders = append(signedHeaders, k)
		}
	}

	return strings.Join(headers, "\n") + "\n", strings.Join(signedHeaders, ";")
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
	_, signedHeaders := s.canonicalHeaders(req.Header)
	scope := s.credentialScope(region, t)

	return "AWS4-HMAC-SHA256 " +
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
