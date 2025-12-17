/*
 * RustFS Go SDK
 * Copyright 2025 RustFS Contributors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package signer 提供 AWS Signature V4/V2 签名功能
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

// Signature and API related constants
const (
	SignV4Algorithm   = "AWS4-HMAC-SHA256"
	Iso8601DateFormat = "20060102T150405Z"
	Yyyymmdd          = "20060102"
)

// SignV4 使用 AWS Signature Version 4 签名 HTTP 请求
func SignV4(req *http.Request, accessKeyID, secretAccessKey, sessionToken, region, service string) *http.Request {
	if accessKeyID == "" || secretAccessKey == "" {
		return req
	}

	t := time.Now().UTC()
	req.Header.Set("X-Amz-Date", t.Format(Iso8601DateFormat))

	if sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", sessionToken)
	}

	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", req.URL.Host)
	}

	// 如果没有设置 Content-SHA256，使用 UNSIGNED-PAYLOAD
	if req.Header.Get("X-Amz-Content-Sha256") == "" {
		req.Header.Set("X-Amz-Content-Sha256", "UNSIGNED-PAYLOAD")
	}

	scope := buildCredentialScope(t, region, service)
	canonicalRequest := buildCanonicalRequest(req)
	stringToSign := buildStringToSign(canonicalRequest, t, scope)
	signingKey := deriveSigningKey(secretAccessKey, t, region, service)
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	signedHeaders := getSignedHeaders(req.Header)
	authorization := SignV4Algorithm + " " +
		"Credential=" + accessKeyID + "/" + scope + ", " +
		"SignedHeaders=" + signedHeaders + ", " +
		"Signature=" + signature

	req.Header.Set("Authorization", authorization)

	return req
}

// SignV4STS 为 STS 请求签名（用于 AssumeRole 等操作）
func SignV4STS(req http.Request, accessKeyID, secretAccessKey, location string) *http.Request {
	region := location
	if region == "" {
		region = "us-east-1"
	}
	return SignV4(&req, accessKeyID, secretAccessKey, "", region, "sts")
}

func buildCredentialScope(t time.Time, region, service string) string {
	return strings.Join([]string{
		t.Format(Yyyymmdd),
		region,
		service,
		"aws4_request",
	}, "/")
}

func buildCanonicalRequest(req *http.Request) string {
	signedHeaders := getSignedHeaders(req.Header)

	var canonicalHeaders strings.Builder
	headers := strings.Split(signedHeaders, ";")
	for _, h := range headers {
		canonicalHeaders.WriteString(h)
		canonicalHeaders.WriteString(":")
		canonicalHeaders.WriteString(strings.TrimSpace(req.Header.Get(h)))
		canonicalHeaders.WriteString("\n")
	}

	encodedPath := req.URL.EscapedPath()
	if encodedPath == "" {
		encodedPath = "/"
	}

	canonicalQuery := req.URL.Query().Encode()

	return strings.Join([]string{
		req.Method,
		encodedPath,
		canonicalQuery,
		canonicalHeaders.String(),
		signedHeaders,
		req.Header.Get("X-Amz-Content-Sha256"),
	}, "\n")
}

func buildStringToSign(canonicalRequest string, t time.Time, scope string) string {
	hash := sha256.Sum256([]byte(canonicalRequest))
	return strings.Join([]string{
		SignV4Algorithm,
		t.Format(Iso8601DateFormat),
		scope,
		hex.EncodeToString(hash[:]),
	}, "\n")
}

func deriveSigningKey(secretAccessKey string, t time.Time, region, service string) []byte {
	kSecret := []byte("AWS4" + secretAccessKey)
	kDate := hmacSHA256(kSecret, []byte(t.Format(Yyyymmdd)))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}

func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func getSignedHeaders(header http.Header) string {
	var headers []string
	for k := range header {
		lowerKey := strings.ToLower(k)
		if lowerKey == "authorization" || lowerKey == "user-agent" {
			continue
		}
		headers = append(headers, lowerKey)
	}

	sort.Strings(headers)
	return strings.Join(headers, ";")
}
