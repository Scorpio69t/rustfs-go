package signer

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Signature and API related constants for V2
const (
	signV2Algorithm = "AWS"
)

// v2ResourceList whitelist of query params to include in V2 signature
// Sorted alphabetically
var v2ResourceList = []string{
	"acl",
	"cors",
	"delete",
	"encryption",
	"lifecycle",
	"location",
	"logging",
	"notification",
	"partNumber",
	"policy",
	"replication",
	"requestPayment",
	"response-cache-control",
	"response-content-disposition",
	"response-content-encoding",
	"response-content-language",
	"response-content-type",
	"response-expires",
	"tagging",
	"torrent",
	"uploadId",
	"uploads",
	"versionId",
	"versioning",
	"versions",
	"website",
}

// V2Signer implements AWS Signature Version 2 signing
type V2Signer struct {
	virtualHost bool
}

// Sign signs request using V2 algorithm
// Authorization = "AWS" + " " + AWSAccessKeyId + ":" + Signature
// Signature = Base64( HMAC-SHA1( YourSecretAccessKeyID, UTF-8-Encoding-Of( StringToSign ) ) )
func (s *V2Signer) Sign(req *http.Request, accessKey, secretKey, sessionToken, region string) *http.Request {
	// Anonymous credentials do not require signing
	if accessKey == "" || secretKey == "" {
		return req
	}

	// Initialize time
	d := time.Now().UTC()

	// Add Date header if missing
	if date := req.Header.Get("Date"); date == "" {
		req.Header.Set("Date", d.Format(http.TimeFormat))
	}

	// Calculate HMAC
	stringToSign := s.stringToSignV2(req)
	hm := hmac.New(sha1.New, []byte(secretKey))
	hm.Write([]byte(stringToSign))

	// Prepare Authorization header
	authHeader := new(bytes.Buffer)
	fmt.Fprintf(authHeader, "%s %s:", signV2Algorithm, accessKey)
	encoder := base64.NewEncoder(base64.StdEncoding, authHeader)
	encoder.Write(hm.Sum(nil))
	encoder.Close()

	// Set Authorization header
	req.Header.Set("Authorization", authHeader.String())

	return req
}

// Presign generates presigned URL using V2 algorithm
// https://${S3_BUCKET}.s3.amazonaws.com/${S3_OBJECT}?AWSAccessKeyId=${S3_ACCESS_KEY}&Expires=${TIMESTAMP}&Signature=${SIGNATURE}
func (s *V2Signer) Presign(req *http.Request, accessKey, secretKey, sessionToken, region string, expires time.Duration) *http.Request {
	// Anonymous credentials do not require signing
	if accessKey == "" || secretKey == "" {
		return req
	}

	d := time.Now().UTC()
	// Compute expiration (Unix timestamp)
	epochExpires := d.Unix() + int64(expires.Seconds())

	// Add Expires header if missing
	if expiresStr := req.Header.Get("Expires"); expiresStr == "" {
		req.Header.Set("Expires", strconv.FormatInt(epochExpires, 10))
	}

	// Get string to sign for presign
	stringToSign := s.preStringToSignV2(req)
	hm := hmac.New(sha1.New, []byte(secretKey))
	hm.Write([]byte(stringToSign))

	// Compute signature
	signature := base64.StdEncoding.EncodeToString(hm.Sum(nil))

	query := req.URL.Query()
	// Handle Google Cloud Storage special case
	if strings.Contains(getHostAddr(req), ".storage.googleapis.com") {
		query.Set("GoogleAccessId", accessKey)
	} else {
		query.Set("AWSAccessKeyId", accessKey)
	}

	// Set Expires query parameter
	query.Set("Expires", strconv.FormatInt(epochExpires, 10))

	// Encode query parameters and save
	req.URL.RawQuery = queryEncode(query)

	// Finally append signature
	req.URL.RawQuery += "&Signature=" + encodePath(signature)

	return req
}

// stringToSignV2 builds string to sign
// StringToSign = HTTP-Verb + "\n" +
//
//	Content-Md5 + "\n" +
//	Content-Type + "\n" +
//	Date + "\n" +
//	CanonicalizedProtocolHeaders +
//	CanonicalizedResource
func (s *V2Signer) stringToSignV2(req *http.Request) string {
	buf := new(bytes.Buffer)
	// Write standard headers
	s.writeSignV2Headers(buf, req)
	// Write canonicalized protocol headers (if any)
	s.writeCanonicalizedHeaders(buf, req)
	// Write canonicalized resource (if any)
	s.writeCanonicalizedResource(buf, req)
	return buf.String()
}

// preStringToSignV2 builds string to sign for presign
func (s *V2Signer) preStringToSignV2(req *http.Request) string {
	buf := new(bytes.Buffer)
	// Write standard headers
	s.writePreSignV2Headers(buf, req)
	// Write canonicalized protocol headers (if any)
	s.writeCanonicalizedHeaders(buf, req)
	// Write canonicalized resource (if any)
	s.writeCanonicalizedResource(buf, req)
	return buf.String()
}

// writeSignV2Headers writes standard headers for V2 signing
func (s *V2Signer) writeSignV2Headers(buf *bytes.Buffer, req *http.Request) {
	buf.WriteString(req.Method + "\n")
	buf.WriteString(req.Header.Get("Content-Md5") + "\n")
	buf.WriteString(req.Header.Get("Content-Type") + "\n")
	buf.WriteString(req.Header.Get("Date") + "\n")
}

// writePreSignV2Headers writes standard headers for V2 presign
func (s *V2Signer) writePreSignV2Headers(buf *bytes.Buffer, req *http.Request) {
	buf.WriteString(req.Method + "\n")
	buf.WriteString(req.Header.Get("Content-Md5") + "\n")
	buf.WriteString(req.Header.Get("Content-Type") + "\n")
	buf.WriteString(req.Header.Get("Expires") + "\n")
}

// writeCanonicalizedHeaders writes canonicalized headers
func (s *V2Signer) writeCanonicalizedHeaders(buf *bytes.Buffer, req *http.Request) {
	var protoHeaders []string
	vals := make(map[string][]string)
	for k, vv := range req.Header {
		// All AMZ headers should be lowercase
		lk := strings.ToLower(k)
		if strings.HasPrefix(lk, "x-amz") {
			protoHeaders = append(protoHeaders, lk)
			vals[lk] = vv
		}
	}
	sort.Strings(protoHeaders)
	for _, k := range protoHeaders {
		buf.WriteString(k)
		buf.WriteByte(':')
		for idx, v := range vals[k] {
			if idx > 0 {
				buf.WriteByte(',')
			}
			buf.WriteString(v)
		}
		buf.WriteByte('\n')
	}
}

// writeCanonicalizedResource writes canonicalized resource
// CanonicalizedResource = [ "/" + Bucket ] +
//
//	<HTTP-Request-URI, from the protocol name up to the query string> +
//	[ subresource, if present ]
func (s *V2Signer) writeCanonicalizedResource(buf *bytes.Buffer, req *http.Request) {
	// Get encoded path
	path := s.encodeURL2Path(req)
	buf.WriteString(path)

	// Handle sub-resources in query parameters
	query := req.URL.Query()
	var resourceList []string
	for _, resource := range v2ResourceList {
		if query.Get(resource) != "" {
			resourceList = append(resourceList, resource+"="+query.Get(resource))
		} else if _, ok := query[resource]; ok {
			resourceList = append(resourceList, resource)
		}
	}

	if len(resourceList) > 0 {
		buf.WriteByte('?')
		buf.WriteString(strings.Join(resourceList, "&"))
	}
}

// encodeURL2Path encodes URL path
func (s *V2Signer) encodeURL2Path(req *http.Request) string {
	if s.virtualHost {
		reqHost := getHostAddr(req)
		dotPos := strings.Index(reqHost, ".")
		if dotPos > -1 {
			bucketName := reqHost[:dotPos]
			path := "/" + bucketName
			path += req.URL.Path
			path = encodePath(path)
			return path
		}
	}
	return encodePath(req.URL.Path)
}
