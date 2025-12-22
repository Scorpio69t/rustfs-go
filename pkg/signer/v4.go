// Package signer internal/signer/v4.go
package signer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	// "fmt"
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

// v4IgnoredHeaders are skipped during signing
// Reference: https://github.com/aws/aws-sdk-js/issues/659#issuecomment-120477258
var v4IgnoredHeaders = map[string]bool{
	"Accept-Encoding": true,
	"Authorization":   true,
	"User-Agent":      true,
}

// V4Signer AWS Signature Version 4 signer
type V4Signer struct{}

// Sign signs a request with Signature V4
func (s *V4Signer) Sign(req *http.Request, accessKey, secretKey, sessionToken, region string) *http.Request {
	// Skip signing when credentials are empty
	if accessKey == "" || secretKey == "" {
		return req
	}

	// Set timestamp
	t := time.Now().UTC()
	req.Header.Set("X-Amz-Date", t.Format(iso8601DateFormat))

	// Attach session token
	if sessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", sessionToken)
	}

	// Set Content-SHA256 header (required by SigV4)
	// Use UNSIGNED-PAYLOAD when it is not provided
	if req.Header.Get("X-Amz-Content-Sha256") == "" {
		req.Header.Set("X-Amz-Content-Sha256", UnsignedPayload)
	}

	// Ensure Host header is present
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", getHostAddr(req))
	}

	// Calculate signature
	signature := s.calculateSignature(req, accessKey, secretKey, region, t)

	// Build Authorization header
	auth := s.buildAuthorizationHeader(req, accessKey, region, signature, t)
	req.Header.Set("Authorization", auth)

	return req
}

// Presign generates a Signature V4 presigned request
// Reference: http://docs.aws.amazon.com/AmazonS3/latest/API/sigv4-query-string-auth.html
func (s *V4Signer) Presign(req *http.Request, accessKey, secretKey, sessionToken, region string, expires time.Duration) *http.Request {
	// Skip signing when credentials are empty
	if accessKey == "" || secretKey == "" {
		return req
	}

	// Initialize timestamp
	t := time.Now().UTC()

	// Build credential string
	credential := s.getCredential(accessKey, region, t)

	// Collect signed headers
	signedHeaders := s.getSignedHeaders(req.Header)

	// Set query parameters
	query := req.URL.Query()
	query.Set("X-Amz-Algorithm", signV4Algorithm)
	query.Set("X-Amz-Date", t.Format(iso8601DateFormat))
	query.Set("X-Amz-Expires", strconv.FormatInt(int64(expires.Seconds()), 10))
	query.Set("X-Amz-SignedHeaders", signedHeaders)
	query.Set("X-Amz-Credential", credential)

	// Include session token when present
	if sessionToken != "" {
		query.Set("X-Amz-Security-Token", sessionToken)
	}

	req.URL.RawQuery = query.Encode()

	// Build canonical request
	canonicalRequest := s.createCanonicalRequest(req)

	// Build string to sign
	stringToSign := s.createStringToSign(canonicalRequest, region, t)

	// Derive signing key
	signingKey := s.deriveSigningKey(secretKey, region, t)

	// Compute signature
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	// Append signature to query params
	req.URL.RawQuery += "&X-Amz-Signature=" + signature

	return req
}

// getCredential builds the credential string
func (s *V4Signer) getCredential(accessKeyID, region string, t time.Time) string {
	scope := s.credentialScope(region, t)
	return accessKeyID + "/" + scope
}

// calculateSignature computes the signature
func (s *V4Signer) calculateSignature(req *http.Request, accessKey, secretKey, region string, t time.Time) string {
	// 1. Create canonical request
	canonicalRequest := s.createCanonicalRequest(req)

	// Debug output
	// fmt.Printf("=== V4 Signature Debug ===\n")
	// fmt.Printf("Method: %s\n", req.Method)
	// fmt.Printf("URL: %s\n", req.URL.String())
	// fmt.Printf("Host Header: %s\n", req.Header.Get("Host"))
	// fmt.Printf("X-Amz-Content-Sha256: %s\n", req.Header.Get("X-Amz-Content-Sha256"))
	// fmt.Printf("Canonical Request:\n%s\n", canonicalRequest)
	// fmt.Printf("=========================\n")

	// 2. Build string to sign
	stringToSign := s.createStringToSign(canonicalRequest, region, t)

	// 3. Compute signature
	signingKey := s.deriveSigningKey(secretKey, region, t)
	signature := hmacSHA256(signingKey, []byte(stringToSign))

	return hex.EncodeToString(signature)
}

// createCanonicalRequest builds the canonical request
// Format: <HTTPMethod>\n<CanonicalURI>\n<CanonicalQueryString>\n<CanonicalHeaders>\n<SignedHeaders>\n<HashedPayload>
func (s *V4Signer) createCanonicalRequest(req *http.Request) string {
	// HTTP Method
	method := req.Method

	// Canonical URI - URL-encode the path
	uri := encodePath(req.URL.Path)
	if uri == "" {
		uri = "/"
	}

	// Canonical Query String - replace + with %20
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

// getHashedPayload returns the payload hash
func (s *V4Signer) getHashedPayload(req *http.Request) string {
	hashedPayload := req.Header.Get("X-Amz-Content-Sha256")
	if hashedPayload == "" {
		// Presign has no payload; use the S3 recommended value
		hashedPayload = UnsignedPayload
	}
	return hashedPayload
}

// getCanonicalHeaders builds canonical headers
func (s *V4Signer) getCanonicalHeaders(req *http.Request) string {
	var headers []string
	vals := make(map[string][]string)

	for k, vv := range req.Header {
		if _, ok := v4IgnoredHeaders[http.CanonicalHeaderKey(k)]; ok {
			continue // header is ignored for signing
		}
		lowerKey := strings.ToLower(k)
		headers = append(headers, lowerKey)
		vals[lowerKey] = vv
	}

	// Ensure Host header is included
	if !headerExists("host", headers) {
		headers = append(headers, "host")
	}
	sort.Strings(headers)

	var buf bytes.Buffer
	// Write headers as <header>:<value> separated by newlines
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

// getSignedHeaders collects signed headers
// Returns lowercase header names sorted and joined by semicolons
func (s *V4Signer) getSignedHeaders(header http.Header) string {
	var headers []string
	for k := range header {
		if _, ok := v4IgnoredHeaders[http.CanonicalHeaderKey(k)]; ok {
			continue // header is ignored for signing
		}
		headers = append(headers, strings.ToLower(k))
	}
	if !headerExists("host", headers) {
		headers = append(headers, "host")
	}
	sort.Strings(headers)
	return strings.Join(headers, ";")
}

// createStringToSign builds the string to sign
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

// credentialScope builds the credential scope
func (s *V4Signer) credentialScope(region string, t time.Time) string {
	return strings.Join([]string{
		t.Format("20060102"),
		region,
		"s3",
		"aws4_request",
	}, "/")
}

// deriveSigningKey derives the signing key
func (s *V4Signer) deriveSigningKey(secretKey, region string, t time.Time) []byte {
	dateKey := hmacSHA256([]byte("AWS4"+secretKey), []byte(t.Format("20060102")))
	regionKey := hmacSHA256(dateKey, []byte(region))
	serviceKey := hmacSHA256(regionKey, []byte("s3"))
	signingKey := hmacSHA256(serviceKey, []byte("aws4_request"))
	return signingKey
}

// buildAuthorizationHeader builds the Authorization header
func (s *V4Signer) buildAuthorizationHeader(req *http.Request, accessKey, region, signature string, t time.Time) string {
	signedHeaders := s.getSignedHeaders(req.Header)
	scope := s.credentialScope(region, t)

	return signV4Algorithm + " " +
		"Credential=" + accessKey + "/" + scope + ", " +
		"SignedHeaders=" + signedHeaders + ", " +
		"Signature=" + signature
}

// hmacSHA256 computes HMAC-SHA256
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// SignV4STS signs STS requests (e.g., AssumeRole)
// Convenience helper dedicated to STS
func SignV4STS(req http.Request, accessKeyID, secretAccessKey, location string) *http.Request {
	region := location
	if region == "" {
		region = "us-east-1"
	}

	if accessKeyID == "" || secretAccessKey == "" {
		return &req
	}

	// Set timestamp
	t := time.Now().UTC()
	req.Header.Set("X-Amz-Date", t.Format(iso8601DateFormat))

	// Ensure Host header exists
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", req.URL.Host)
	}

	// Set Content-SHA256 header if missing
	if req.Header.Get("X-Amz-Content-Sha256") == "" {
		req.Header.Set("X-Amz-Content-Sha256", UnsignedPayload)
	}

	// Use V4Signer to sign for the STS service
	signer := &V4Signer{}
	signature := signer.calculateSignature(&req, accessKeyID, secretAccessKey, region, t)
	auth := signer.buildAuthorizationHeader(&req, accessKeyID, region, signature, t)
	req.Header.Set("Authorization", auth)

	return &req
}
