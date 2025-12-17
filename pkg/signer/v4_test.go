// Package signer internal/signer/v4_test.go
package signer

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestV4Signer_Sign(t *testing.T) {
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
				"Content-Type":         "text/plain",
				"X-Amz-Content-Sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			},
			accessKey:      "AKIAIOSFODNN7EXAMPLE",
			secretKey:      "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			region:         "us-east-1",
			wantAuthHeader: true,
			wantDateHeader: true,
		},
		{
			name:           "Request with session token",
			method:         "GET",
			urlStr:         "https://s3.amazonaws.com/examplebucket/test.txt",
			headers:        map[string]string{},
			accessKey:      "AKIAIOSFODNN7EXAMPLE",
			secretKey:      "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			sessionToken:   "AQoDYXdzEJr1K...SessionToken...yz",
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

			signer := &V4Signer{}
			signedReq := signer.Sign(req, tt.accessKey, tt.secretKey, tt.sessionToken, tt.region)

			// 检查 Authorization 头
			if tt.wantAuthHeader {
				authHeader := signedReq.Header.Get("Authorization")
				if authHeader == "" {
					t.Error("Expected Authorization header, got none")
				}
				if !strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256") {
					t.Errorf("Authorization header should start with AWS4-HMAC-SHA256, got: %s", authHeader)
				}
				if !strings.Contains(authHeader, "Credential=") {
					t.Error("Authorization header should contain Credential=")
				}
				if !strings.Contains(authHeader, "SignedHeaders=") {
					t.Error("Authorization header should contain SignedHeaders=")
				}
				if !strings.Contains(authHeader, "Signature=") {
					t.Error("Authorization header should contain Signature=")
				}
			} else {
				if signedReq.Header.Get("Authorization") != "" {
					t.Error("Expected no Authorization header for anonymous request")
				}
			}

			// 检查 X-Amz-Date 头
			if tt.wantDateHeader {
				dateHeader := signedReq.Header.Get("X-Amz-Date")
				if dateHeader == "" {
					t.Error("Expected X-Amz-Date header, got none")
				}
				// 验证日期格式
				_, err := time.Parse(iso8601DateFormat, dateHeader)
				if err != nil {
					t.Errorf("X-Amz-Date header has invalid format: %v", err)
				}
			}

			// 检查 Session Token
			if tt.sessionToken != "" {
				tokenHeader := signedReq.Header.Get("X-Amz-Security-Token")
				if tokenHeader != tt.sessionToken {
					t.Errorf("Expected session token %s, got %s", tt.sessionToken, tokenHeader)
				}
			}
		})
	}
}

func TestV4Signer_Presign(t *testing.T) {
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
				"X-Amz-Algorithm",
				"X-Amz-Credential",
				"X-Amz-Date",
				"X-Amz-Expires",
				"X-Amz-SignedHeaders",
				"X-Amz-Signature",
			},
		},
		{
			name:         "Presign with session token",
			method:       "GET",
			urlStr:       "https://s3.amazonaws.com/examplebucket/test.txt",
			accessKey:    "AKIAIOSFODNN7EXAMPLE",
			secretKey:    "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			sessionToken: "SessionToken123",
			region:       "us-east-1",
			expires:      time.Hour,
			wantParams: []string{
				"X-Amz-Algorithm",
				"X-Amz-Credential",
				"X-Amz-Date",
				"X-Amz-Expires",
				"X-Amz-SignedHeaders",
				"X-Amz-Security-Token",
				"X-Amz-Signature",
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

			signer := &V4Signer{}
			presignedReq := signer.Presign(req, tt.accessKey, tt.secretKey, tt.sessionToken, tt.region, tt.expires)

			query := presignedReq.URL.Query()

			for _, param := range tt.wantParams {
				if query.Get(param) == "" {
					t.Errorf("Expected query parameter %s, got none", param)
				}
			}

			// 验证算法
			if tt.accessKey != "" && query.Get("X-Amz-Algorithm") != "AWS4-HMAC-SHA256" {
				t.Errorf("Expected X-Amz-Algorithm=AWS4-HMAC-SHA256, got %s", query.Get("X-Amz-Algorithm"))
			}

			// 验证过期时间
			if tt.accessKey != "" {
				expiresParam := query.Get("X-Amz-Expires")
				if expiresParam == "" {
					t.Error("Expected X-Amz-Expires parameter")
				}
			}

			// 验证 session token
			if tt.sessionToken != "" {
				tokenParam := query.Get("X-Amz-Security-Token")
				if tokenParam != tt.sessionToken {
					t.Errorf("Expected session token %s in query, got %s", tt.sessionToken, tokenParam)
				}
			}
		})
	}
}

func TestV4Signer_GetCanonicalHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{
			name: "Basic headers",
			headers: map[string]string{
				"Host":         "s3.amazonaws.com",
				"Content-Type": "text/plain",
			},
			want: "content-type:text/plain\nhost:s3.amazonaws.com\n",
		},
		{
			name: "Headers with spaces",
			headers: map[string]string{
				"Host":            "s3.amazonaws.com",
				"X-Amz-Meta-Test": "  value  with  spaces  ",
			},
			want: "host:s3.amazonaws.com\nx-amz-meta-test:value with spaces\n",
		},
		{
			name: "Ignored headers",
			headers: map[string]string{
				"Host":            "s3.amazonaws.com",
				"Authorization":   "should be ignored",
				"User-Agent":      "should be ignored",
				"Accept-Encoding": "should be ignored",
			},
			want: "host:s3.amazonaws.com\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "https://s3.amazonaws.com/bucket/key", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			signer := &V4Signer{}
			got := signer.getCanonicalHeaders(req)

			if got != tt.want {
				t.Errorf("getCanonicalHeaders() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestV4Signer_GetSignedHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		want    string
	}{
		{
			name: "Basic headers",
			headers: map[string]string{
				"Host":         "s3.amazonaws.com",
				"Content-Type": "text/plain",
			},
			want: "content-type;host",
		},
		{
			name: "With X-Amz headers",
			headers: map[string]string{
				"Host":                 "s3.amazonaws.com",
				"X-Amz-Date":           "20230101T000000Z",
				"X-Amz-Content-Sha256": "hash",
			},
			want: "host;x-amz-content-sha256;x-amz-date",
		},
		{
			name: "Ignored headers excluded",
			headers: map[string]string{
				"Host":            "s3.amazonaws.com",
				"User-Agent":      "test",
				"Authorization":   "test",
				"Accept-Encoding": "gzip",
			},
			want: "host",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := http.Header{}
			for k, v := range tt.headers {
				header.Set(k, v)
			}

			signer := &V4Signer{}
			got := signer.getSignedHeaders(header)

			if got != tt.want {
				t.Errorf("getSignedHeaders() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestV4Signer_CredentialScope(t *testing.T) {
	signer := &V4Signer{}
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name   string
		region string
		want   string
	}{
		{
			name:   "US East 1",
			region: "us-east-1",
			want:   "20230101/us-east-1/s3/aws4_request",
		},
		{
			name:   "EU West 1",
			region: "eu-west-1",
			want:   "20230101/eu-west-1/s3/aws4_request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := signer.credentialScope(tt.region, testTime)
			if got != tt.want {
				t.Errorf("credentialScope() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEncodePath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "Simple path",
			path: "/bucket/key",
			want: "/bucket/key",
		},
		{
			name: "Path with spaces",
			path: "/bucket/my file.txt",
			want: "/bucket/my%20file.txt",
		},
		{
			name: "Path with special characters",
			path: "/bucket/file name.txt",
			want: "/bucket/file%20name.txt",
		},
		{
			name: "Empty path",
			path: "",
			want: "/",
		},
		{
			name: "Root path",
			path: "/",
			want: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := encodePath(tt.path)
			if got != tt.want {
				t.Errorf("encodePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestQueryEncode(t *testing.T) {
	tests := []struct {
		name  string
		query url.Values
		want  string
	}{
		{
			name: "Simple query",
			query: url.Values{
				"key": []string{"value"},
			},
			want: "key=value",
		},
		{
			name: "Multiple values",
			query: url.Values{
				"key": []string{"value1", "value2"},
			},
			want: "key=value1&key=value2",
		},
		{
			name: "Sorted keys",
			query: url.Values{
				"z": []string{"last"},
				"a": []string{"first"},
				"m": []string{"middle"},
			},
			want: "a=first&m=middle&z=last",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := queryEncode(tt.query)
			if got != tt.want {
				t.Errorf("queryEncode() = %q, want %q", got, tt.want)
			}
		})
	}
}
