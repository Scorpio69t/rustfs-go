// Package object object/post_policy_test.go
package object

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Scorpio69t/rustfs-go/pkg/policy"
)

func TestPresignedPostPolicy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET request, got %s", r.Method)
		}
		if r.URL.Query().Has("location") {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><LocationConstraint>us-east-1</LocationConstraint>`)); err != nil {
				t.Fatalf("Failed to write location response: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)

	postPolicy := policy.NewPostPolicy()
	if err := postPolicy.SetExpires(time.Now().UTC().Add(time.Hour)); err != nil {
		t.Fatalf("SetExpires() error = %v", err)
	}
	if err := postPolicy.SetBucket("test-bucket"); err != nil {
		t.Fatalf("SetBucket() error = %v", err)
	}
	if err := postPolicy.SetKey("test-key"); err != nil {
		t.Fatalf("SetKey() error = %v", err)
	}

	postURL, formData, err := service.PresignedPostPolicy(context.Background(), postPolicy)
	if err != nil {
		t.Fatalf("PresignedPostPolicy() error = %v", err)
	}
	if postURL == nil {
		t.Fatal("PresignedPostPolicy() returned nil URL")
	}

	serverURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("Parse server URL error = %v", err)
	}

	if postURL.Host != serverURL.Host {
		t.Fatalf("post URL host = %q, want %q", postURL.Host, serverURL.Host)
	}
	if !strings.HasSuffix(postURL.Path, "/test-bucket/") {
		t.Fatalf("post URL path = %q, want suffix %q", postURL.Path, "/test-bucket/")
	}

	wantFields := []string{
		"policy",
		"x-amz-algorithm",
		"x-amz-credential",
		"x-amz-date",
		"x-amz-signature",
		"bucket",
		"key",
	}
	for _, field := range wantFields {
		if formData[field] == "" {
			t.Fatalf("formData[%q] is empty", field)
		}
	}
	if formData["x-amz-algorithm"] != postPolicyAlgorithm {
		t.Fatalf("x-amz-algorithm = %q, want %q", formData["x-amz-algorithm"], postPolicyAlgorithm)
	}
}
