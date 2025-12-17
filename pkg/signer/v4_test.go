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

package signer

import (
	"net/http"
	"strings"
	"testing"
)

func TestSignV4(t *testing.T) {
	tests := []struct {
		name            string
		accessKeyID     string
		secretAccessKey string
		sessionToken    string
		region          string
		service         string
		wantAuth        bool
	}{
		{
			name:            "Valid credentials",
			accessKeyID:     "AKIAIOSFODNN7EXAMPLE",
			secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			sessionToken:    "",
			region:          "us-east-1",
			service:         "s3",
			wantAuth:        true,
		},
		{
			name:            "Empty credentials",
			accessKeyID:     "",
			secretAccessKey: "",
			sessionToken:    "",
			region:          "us-east-1",
			service:         "s3",
			wantAuth:        false,
		},
		{
			name:            "With session token",
			accessKeyID:     "AKIAIOSFODNN7EXAMPLE",
			secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			sessionToken:    "session-token",
			region:          "us-west-2",
			service:         "s3",
			wantAuth:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "https://s3.amazonaws.com/bucket/key", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			signedReq := SignV4(req, tt.accessKeyID, tt.secretAccessKey, tt.sessionToken, tt.region, tt.service)

			if tt.wantAuth {
				if signedReq.Header.Get("Authorization") == "" {
					t.Error("Expected Authorization header, got none")
				}
				if signedReq.Header.Get("X-Amz-Date") == "" {
					t.Error("Expected X-Amz-Date header, got none")
				}
				if signedReq.Header.Get("X-Amz-Content-Sha256") == "" {
					t.Error("Expected X-Amz-Content-Sha256 header, got none")
				}
				if tt.sessionToken != "" && signedReq.Header.Get("X-Amz-Security-Token") == "" {
					t.Error("Expected X-Amz-Security-Token header, got none")
				}
			} else {
				if signedReq.Header.Get("Authorization") != "" {
					t.Error("Expected no Authorization header for empty credentials")
				}
			}
		})
	}
}

func TestSignV4STS(t *testing.T) {
	tests := []struct {
		name            string
		accessKeyID     string
		secretAccessKey string
		location        string
		wantAuth        bool
	}{
		{
			name:            "Valid STS credentials",
			accessKeyID:     "AKIAIOSFODNN7EXAMPLE",
			secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			location:        "us-east-1",
			wantAuth:        true,
		},
		{
			name:            "Empty location defaults to us-east-1",
			accessKeyID:     "AKIAIOSFODNN7EXAMPLE",
			secretAccessKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
			location:        "",
			wantAuth:        true,
		},
		{
			name:            "Empty credentials",
			accessKeyID:     "",
			secretAccessKey: "",
			location:        "us-east-1",
			wantAuth:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "https://sts.amazonaws.com/", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("X-Amz-Content-Sha256", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")

			signedReq := SignV4STS(*req, tt.accessKeyID, tt.secretAccessKey, tt.location)

			if tt.wantAuth {
				if signedReq.Header.Get("Authorization") == "" {
					t.Error("Expected Authorization header, got none")
				}
				if signedReq.Header.Get("X-Amz-Date") == "" {
					t.Error("Expected X-Amz-Date header, got none")
				}
			} else {
				if signedReq.Header.Get("Authorization") != "" {
					t.Error("Expected no Authorization header for empty credentials")
				}
			}
		})
	}
}

func TestGetSignedHeaders(t *testing.T) {
	header := http.Header{}
	header.Set("Host", "s3.amazonaws.com")
	header.Set("X-Amz-Date", "20150830T123600Z")
	header.Set("Authorization", "should-be-excluded")
	header.Set("User-Agent", "should-be-excluded")

	signed := getSignedHeaders(header)

	// Authorization and User-Agent should be excluded
	if strings.Contains(signed, "authorization") {
		t.Error("Authorization header should be excluded")
	}
	if strings.Contains(signed, "user-agent") {
		t.Error("User-Agent header should be excluded")
	}

	// Host and X-Amz-Date should be included
	if !strings.Contains(signed, "host") {
		t.Error("Host header should be included")
	}
	if !strings.Contains(signed, "x-amz-date") {
		t.Error("X-Amz-Date header should be included")
	}
}
