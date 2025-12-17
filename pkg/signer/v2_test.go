// Package signer internal/signer/v2_test.go
package signer

import (
	"net/http"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestV2Signer_Sign(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		urlStr         string
		headers        map[string]string
		accessKey      string
		secretKey      string
		sessionToken   string
		region         string
		wantAuthHeader bool
		wantDateHeader bool
	}{
		{
			name:           "Basic GET request",
			method:         "GET",
			urlStr:         "https://s3.amazonaws.com/examplebucket/test.txt",
			headers:        map[string]string{},
			accessKey:      "AKIAIOSFODNN7EXAMPLE",
			secretKey:      "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			region:         "us-east-1",
			wantAuthHeader: true,
			wantDateHeader: true,
		},
		{
			name:   "PUT request with content",
			method: "PUT",
			urlStr: "https://s3.amazonaws.com/examplebucket/test.txt",
			headers: map[string]string{
				"Content-Type": "text/plain",
				"Content-Md5":  "rL0Y20zC+Fzt72VPzMSk2A==",
			},
			accessKey:      "AKIAIOSFODNN7EXAMPLE",
			secretKey:      "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			region:         "us-east-1",
			wantAuthHeader: true,
			wantDateHeader: true,
		},
		{
			name:   "Request with X-Amz headers",
			method: "GET",
			urlStr: "https://s3.amazonaws.com/examplebucket/test.txt",
			headers: map[string]string{
				"X-Amz-Meta-Test": "test-value",
				"X-Amz-Date":      "Mon, 02 Jan 2006 15:04:05 GMT",
			},
			accessKey:      "AKIAIOSFODNN7EXAMPLE",
			secretKey:      "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			region:         "us-east-1",
			wantAuthHeader: true,
			wantDateHeader: true,
		},
		{
			name:           "Anonymous request (no credentials)",
			method:         "GET",
			urlStr:         "https://s3.amazonaws.com/examplebucket/test.txt",
			headers:        map[string]string{},
			accessKey:      "",
			secretKey:      "",
			region:         "us-east-1",
			wantAuthHeader: false,
			wantDateHeader: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.urlStr, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			signer := &V2Signer{}
			signedReq := signer.Sign(req, tt.accessKey, tt.secretKey, tt.sessionToken, tt.region)

			// 检查 Authorization 头
			if tt.wantAuthHeader {
				authHeader := signedReq.Header.Get("Authorization")
				if authHeader == "" {
					t.Error("Expected Authorization header, got none")
				}
				if !strings.HasPrefix(authHeader, "AWS ") {
					t.Errorf("Authorization header should start with 'AWS ', got: %s", authHeader)
				}
				if !strings.Contains(authHeader, ":") {
					t.Error("Authorization header should contain ':' separator")
				}
			} else {
				if signedReq.Header.Get("Authorization") != "" {
					t.Error("Expected no Authorization header for anonymous request")
				}
			}

			// 检查 Date 头
			if tt.wantDateHeader {
				dateHeader := signedReq.Header.Get("Date")
				if dateHeader == "" {
					t.Error("Expected Date header, got none")
				}
				// 验证日期格式
				_, err := time.Parse(http.TimeFormat, dateHeader)
				if err != nil {
					t.Errorf("Date header has invalid format: %v", err)
				}
			}
		})
	}
}

func TestV2Signer_Presign(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		urlStr       string
		accessKey    string
		secretKey    string
		sessionToken string
		region       string
		expires      time.Duration
		wantParams   []string
	}{
		{
			name:      "Presign GET request",
			method:    "GET",
			urlStr:    "https://s3.amazonaws.com/examplebucket/test.txt",
			accessKey: "AKIAIOSFODNN7EXAMPLE",
			secretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			region:    "us-east-1",
			expires:   time.Hour,
			wantParams: []string{
				"AWSAccessKeyId",
				"Expires",
				"Signature",
			},
		},
		{
			name:      "Presign with short expiry",
			method:    "GET",
			urlStr:    "https://s3.amazonaws.com/examplebucket/test.txt",
			accessKey: "AKIAIOSFODNN7EXAMPLE",
			secretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			region:    "us-east-1",
			expires:   5 * time.Minute,
			wantParams: []string{
				"AWSAccessKeyId",
				"Expires",
				"Signature",
			},
		},
		{
			name:       "Anonymous presign (no credentials)",
			method:     "GET",
			urlStr:     "https://s3.amazonaws.com/examplebucket/test.txt",
			accessKey:  "",
			secretKey:  "",
			region:     "us-east-1",
			expires:    time.Hour,
			wantParams: []string{}, // 不应该有任何签名参数
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.urlStr, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			signer := &V2Signer{}
			presignedReq := signer.Presign(req, tt.accessKey, tt.secretKey, tt.sessionToken, tt.region, tt.expires)

			query := presignedReq.URL.Query()

			for _, param := range tt.wantParams {
				if query.Get(param) == "" {
					t.Errorf("Expected query parameter %s, got none", param)
				}
			}

			// 验证 access key
			if tt.accessKey != "" {
				accessKeyParam := query.Get("AWSAccessKeyId")
				if accessKeyParam != tt.accessKey {
					t.Errorf("Expected AWSAccessKeyId=%s, got %s", tt.accessKey, accessKeyParam)
				}
			}

			// 验证过期时间存在
			if tt.accessKey != "" {
				expiresParam := query.Get("Expires")
				if expiresParam == "" {
					t.Error("Expected Expires parameter")
				}
			}

			// 验证签名存在
			if tt.accessKey != "" {
				signatureParam := query.Get("Signature")
				if signatureParam == "" {
					t.Error("Expected Signature parameter")
				}
			}
		})
	}
}

func TestV2Signer_WriteCanonicalizedHeaders(t *testing.T) {
	tests := []struct {
		name         string
		headers      map[string]string
		wantContains []string // 期望包含的头部
	}{
		{
			name: "No X-Amz headers",
			headers: map[string]string{
				"Host":         "s3.amazonaws.com",
				"Content-Type": "text/plain",
			},
			wantContains: []string{}, // 没有 x-amz 头部
		},
		{
			name: "Single X-Amz header",
			headers: map[string]string{
				"X-Amz-Date": "20230101T000000Z",
			},
			wantContains: []string{"x-amz-date:20230101T000000Z"},
		},
		{
			name: "Multiple X-Amz headers (sorted)",
			headers: map[string]string{
				"X-Amz-Meta-Test": "value1",
				"X-Amz-Date":      "20230101T000000Z",
				"X-Amz-ACL":       "public-read",
			},
			wantContains: []string{
				"x-amz-acl:public-read",
				"x-amz-date:20230101T000000Z",
				"x-amz-meta-test:value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "https://s3.amazonaws.com/bucket/key", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			signer := &V2Signer{}
			stringToSign := signer.stringToSignV2(req)

			// 检查是否包含期望的头部
			for _, want := range tt.wantContains {
				if !strings.Contains(stringToSign, want) {
					t.Errorf("stringToSign should contain %q, got: %s", want, stringToSign)
				}
			}
		})
	}
}

func TestV2Signer_WriteCanonicalizedResource(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		query        map[string]string
		wantContains string
	}{
		{
			name:         "Simple path",
			path:         "/bucket/key",
			query:        map[string]string{},
			wantContains: "/bucket/key",
		},
		{
			name: "Path with ACL subresource",
			path: "/bucket/key",
			query: map[string]string{
				"acl": "",
			},
			wantContains: "/bucket/key?acl",
		},
		{
			name: "Path with versioning",
			path: "/bucket/key",
			query: map[string]string{
				"versionId": "abc123",
			},
			wantContains: "/bucket/key?versionId=abc123",
		},
		{
			name: "Path with partNumber",
			path: "/bucket/key",
			query: map[string]string{
				"partNumber": "1",
			},
			wantContains: "/bucket/key?partNumber=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urlStr := "https://s3.amazonaws.com" + tt.path
			if len(tt.query) > 0 {
				urlStr += "?"
				var parts []string
				for k, v := range tt.query {
					if v == "" {
						parts = append(parts, k)
					} else {
						parts = append(parts, k+"="+v)
					}
				}
				urlStr += strings.Join(parts, "&")
			}

			req, _ := http.NewRequest("GET", urlStr, nil)

			signer := &V2Signer{}
			stringToSign := signer.stringToSignV2(req)

			if !strings.Contains(stringToSign, tt.wantContains) {
				t.Errorf("stringToSign should contain %q, got: %s", tt.wantContains, stringToSign)
			}
		})
	}
}

func TestV2ResourceListSorting(t *testing.T) {
	// 测试 v2ResourceList 是否已正确排序
	sortedList := make([]string, len(v2ResourceList))
	copy(sortedList, v2ResourceList)
	sort.Strings(sortedList)

	for i := 0; i < len(v2ResourceList); i++ {
		if v2ResourceList[i] != sortedList[i] {
			t.Errorf("v2ResourceList[%d] = %q, expected %q (list is not sorted)",
				i, v2ResourceList[i], sortedList[i])
		}
	}
}

func TestV2Signer_GoogleCloudStorage(t *testing.T) {
	// 测试 Google Cloud Storage 的特殊处理
	req, _ := http.NewRequest("GET", "https://bucket.storage.googleapis.com/key", nil)

	signer := &V2Signer{}
	presignedReq := signer.Presign(req, "access-key", "secret-key", "", "us-east-1", time.Hour)

	query := presignedReq.URL.Query()

	// 应该使用 GoogleAccessId 而不是 AWSAccessKeyId
	if query.Get("GoogleAccessId") == "" {
		t.Error("Expected GoogleAccessId parameter for Google Cloud Storage")
	}
	if query.Get("AWSAccessKeyId") != "" {
		t.Error("Should not have AWSAccessKeyId for Google Cloud Storage")
	}
}

func TestSignV4TrimAll(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "No spaces",
			input: "test",
			want:  "test",
		},
		{
			name:  "Leading space",
			input: " test",
			want:  "test",
		},
		{
			name:  "Trailing space",
			input: "test ",
			want:  "test",
		},
		{
			name:  "Multiple spaces",
			input: "test   value   here",
			want:  "test value here",
		},
		{
			name:  "Mixed whitespace",
			input: "  test\t\tvalue  \n here  ",
			want:  "test value here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := signV4TrimAll(tt.input)
			if got != tt.want {
				t.Errorf("signV4TrimAll(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetHostAddr(t *testing.T) {
	tests := []struct {
		name       string
		reqHost    string
		headerHost string
		urlHost    string
		want       string
	}{
		{
			name:       "Use header host",
			reqHost:    "",
			headerHost: "header.example.com",
			urlHost:    "url.example.com",
			want:       "header.example.com",
		},
		{
			name:       "Use req.Host",
			reqHost:    "req.example.com",
			headerHost: "",
			urlHost:    "url.example.com",
			want:       "req.example.com",
		},
		{
			name:       "Use URL host",
			reqHost:    "",
			headerHost: "",
			urlHost:    "url.example.com",
			want:       "url.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "https://"+tt.urlHost+"/path", nil)
			if tt.headerHost != "" {
				req.Header.Set("Host", tt.headerHost)
			}
			if tt.reqHost != "" {
				req.Host = tt.reqHost
			}

			got := getHostAddr(req)
			if got != tt.want {
				t.Errorf("getHostAddr() = %q, want %q", got, tt.want)
			}
		})
	}
}
