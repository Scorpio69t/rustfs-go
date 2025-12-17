// Package s3signer provides S3 signature v4 implementation
package s3signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const (
	iso8601Format = "20060102T150405Z"
	yyyymmdd      = "20060102"
)

// SignV4 - sign request with AWS Signature Version 4
func SignV4(req *http.Request, accessKeyID, secretAccessKey, region, service string, t time.Time) error {
	// Get credentials
	if accessKeyID == "" || secretAccessKey == "" {
		return fmt.Errorf("access key ID and secret access key are required")
	}

	// Set date header
	amzDate := t.UTC().Format(iso8601Format)
	dateStamp := t.UTC().Format(yyyymmdd)

	req.Header.Set("X-Amz-Date", amzDate)

	// Create canonical request
	canonicalRequest := buildCanonicalRequest(req)

	// Create string to sign
	stringToSign := buildStringToSign(amzDate, dateStamp, region, service, canonicalRequest)

	// Calculate signature
	signature := calculateSignature(secretAccessKey, dateStamp, region, service, stringToSign)

	// Create authorization header
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service)
	authorization := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKeyID, credentialScope, getSignedHeaders(req), signature)

	req.Header.Set("Authorization", authorization)

	return nil
}

// buildCanonicalRequest builds the canonical request string
func buildCanonicalRequest(req *http.Request) string {
	// Method
	method := req.Method

	// URI
	uri := req.URL.EscapedPath()
	if uri == "" {
		uri = "/"
	}

	// Query string
	query := req.URL.Query()
	var queryKeys []string
	for k := range query {
		queryKeys = append(queryKeys, k)
	}
	sort.Strings(queryKeys)
	var queryParts []string
	for _, k := range queryKeys {
		if k == "X-Amz-Signature" {
			continue
		}
		v := query.Get(k)
		if v == "" {
			queryParts = append(queryParts, url.QueryEscape(k))
		} else {
			queryParts = append(queryParts, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	queryString := strings.Join(queryParts, "&")

	// Headers
	headers := make(map[string]string)
	for k, v := range req.Header {
		lowerKey := strings.ToLower(k)
		if lowerKey == "host" || strings.HasPrefix(lowerKey, "x-amz-") {
			headers[lowerKey] = strings.TrimSpace(strings.Join(v, ","))
		}
	}
	// Add host header if not present
	if _, ok := headers["host"]; !ok {
		headers["host"] = req.Host
		if headers["host"] == "" {
			headers["host"] = req.URL.Host
		}
	}

	var headerKeys []string
	for k := range headers {
		headerKeys = append(headerKeys, k)
	}
	sort.Strings(headerKeys)

	var canonicalHeaders []string
	for _, k := range headerKeys {
		canonicalHeaders = append(canonicalHeaders, k+":"+headers[k])
	}
	canonicalHeadersStr := strings.Join(canonicalHeaders, "\n") + "\n"

	signedHeaders := strings.Join(headerKeys, ";")

	// Payload hash
	payloadHash := "UNSIGNED-PAYLOAD"

	// Build canonical request
	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s%s\n%s\n",
		method,
		uri,
		queryString,
		canonicalHeadersStr,
		signedHeaders,
		payloadHash)

	return canonicalRequest
}

// buildStringToSign builds the string to sign
func buildStringToSign(amzDate, dateStamp, region, service, canonicalRequest string) string {
	algorithm := "AWS4-HMAC-SHA256"
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", dateStamp, region, service)

	hasher := sha256.New()
	hasher.Write([]byte(canonicalRequest))
	canonicalRequestHash := hex.EncodeToString(hasher.Sum(nil))

	stringToSign := fmt.Sprintf("%s\n%s\n%s\n%s",
		algorithm,
		amzDate,
		credentialScope,
		canonicalRequestHash)

	return stringToSign
}

// calculateSignature calculates the signature
func calculateSignature(secretAccessKey, dateStamp, region, service, stringToSign string) string {
	kDate := hmacSHA256([]byte("AWS4"+secretAccessKey), dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "aws4_request")
	signature := hmacSHA256(kSigning, stringToSign)
	return hex.EncodeToString(signature)
}

// hmacSHA256 calculates HMAC-SHA256
func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// getSignedHeaders returns the signed headers string
func getSignedHeaders(req *http.Request) string {
	headers := make(map[string]bool)
	for k := range req.Header {
		lowerKey := strings.ToLower(k)
		if lowerKey == "host" || strings.HasPrefix(lowerKey, "x-amz-") {
			headers[lowerKey] = true
		}
	}
	// Add host if not present
	if _, ok := headers["host"]; !ok {
		headers["host"] = true
	}

	var headerKeys []string
	for k := range headers {
		headerKeys = append(headerKeys, k)
	}
	sort.Strings(headerKeys)
	return strings.Join(headerKeys, ";")
}

// SignV4Presigned - sign request with AWS Signature Version 4 for presigned URLs
func SignV4Presigned(req *http.Request, accessKeyID, secretAccessKey, region, service string, expiry time.Time) error {
	if accessKeyID == "" || secretAccessKey == "" {
		return fmt.Errorf("access key ID and secret access key are required")
	}

	// Set expiry in query string
	amzDate := expiry.UTC().Format(iso8601Format)
	dateStamp := expiry.UTC().Format(yyyymmdd)

	query := req.URL.Query()
	query.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	query.Set("X-Amz-Credential", fmt.Sprintf("%s/%s/%s/%s/aws4_request", accessKeyID, dateStamp, region, service))
	query.Set("X-Amz-Date", amzDate)
	query.Set("X-Amz-Expires", fmt.Sprintf("%d", int(expiry.Sub(time.Now()).Seconds())))
	query.Set("X-Amz-SignedHeaders", getSignedHeaders(req))
	req.URL.RawQuery = query.Encode()

	// Create canonical request
	canonicalRequest := buildCanonicalRequestPresigned(req)

	// Create string to sign
	stringToSign := buildStringToSign(amzDate, dateStamp, region, service, canonicalRequest)

	// Calculate signature
	signature := calculateSignature(secretAccessKey, dateStamp, region, service, stringToSign)

	// Add signature to query
	query.Set("X-Amz-Signature", signature)
	req.URL.RawQuery = query.Encode()

	return nil
}

// buildCanonicalRequestPresigned builds the canonical request string for presigned URLs
func buildCanonicalRequestPresigned(req *http.Request) string {
	method := req.Method
	uri := req.URL.EscapedPath()
	if uri == "" {
		uri = "/"
	}

	// Query string (already sorted by URL encoding)
	queryString := req.URL.RawQuery

	// Headers
	headers := make(map[string]string)
	for k, v := range req.Header {
		lowerKey := strings.ToLower(k)
		if lowerKey == "host" || strings.HasPrefix(lowerKey, "x-amz-") {
			headers[lowerKey] = strings.TrimSpace(strings.Join(v, ","))
		}
	}
	if _, ok := headers["host"]; !ok {
		headers["host"] = req.Host
		if headers["host"] == "" {
			headers["host"] = req.URL.Host
		}
	}

	var headerKeys []string
	for k := range headers {
		headerKeys = append(headerKeys, k)
	}
	sort.Strings(headerKeys)

	var canonicalHeaders []string
	for _, k := range headerKeys {
		canonicalHeaders = append(canonicalHeaders, k+":"+headers[k])
	}
	canonicalHeadersStr := strings.Join(canonicalHeaders, "\n") + "\n"

	signedHeaders := strings.Join(headerKeys, ";")
	payloadHash := "UNSIGNED-PAYLOAD"

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s%s\n%s\n",
		method,
		uri,
		queryString,
		canonicalHeadersStr,
		signedHeaders,
		payloadHash)

	return canonicalRequest
}

// SignPolicy - sign policy string for POST policy
func SignPolicy(policy, secretAccessKey string) string {
	hash := hmacSHA256([]byte(secretAccessKey), policy)
	return hex.EncodeToString(hash)
}
