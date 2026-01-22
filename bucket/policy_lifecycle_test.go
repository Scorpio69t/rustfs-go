// Package bucket bucket/policy_lifecycle_test.go
package bucket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPolicyAndLifecycle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.RawQuery, "policy"):
			switch r.Method {
			case http.MethodPut:
				w.WriteHeader(http.StatusNoContent)
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(`{"Version":"2012-10-17"}`)); err != nil {
					t.Fatalf("Failed to write policy response: %v", err)
				}
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		case strings.Contains(r.URL.RawQuery, "lifecycle"):
			switch r.Method {
			case http.MethodPut:
				w.WriteHeader(http.StatusOK)
			case http.MethodGet:
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(`<LifecycleConfiguration/>`)); err != nil {
					t.Fatalf("Failed to write lifecycle response: %v", err)
				}
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer server.Close()

	service := createTestService(t, server)
	ctx := context.Background()

	if err := service.SetPolicy(ctx, "bucket", "{}"); err != nil {
		t.Fatalf("SetPolicy() error = %v", err)
	}
	if policy, err := service.GetPolicy(ctx, "bucket"); err != nil || policy == "" {
		t.Fatalf("GetPolicy() error=%v policy=%s", err, policy)
	}
	if err := service.DeletePolicy(ctx, "bucket"); err != nil {
		t.Fatalf("DeletePolicy() error = %v", err)
	}

	if err := service.SetLifecycle(ctx, "bucket", []byte("<xml/>")); err != nil {
		t.Fatalf("SetLifecycle() error = %v", err)
	}
	if cfg, err := service.GetLifecycle(ctx, "bucket"); err != nil || len(cfg) == 0 {
		t.Fatalf("GetLifecycle() error=%v cfg=%s", err, string(cfg))
	}
	if err := service.DeleteLifecycle(ctx, "bucket"); err != nil {
		t.Fatalf("DeleteLifecycle() error = %v", err)
	}
}
