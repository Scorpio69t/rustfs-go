// Package bucket bucket/config_test.go
package bucket

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Scorpio69t/rustfs-go/types"
)

func TestSetAndGetVersioning(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			if _, ok := r.URL.Query()["versioning"]; !ok {
				t.Errorf("expected versioning query")
			}
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "<Status>Enabled</Status>") {
				t.Errorf("expected status in body, got %s", string(body))
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<VersioningConfiguration><Status>Enabled</Status></VersioningConfiguration>`))
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer server.Close()

	service := createTestService(t, server)

	err := service.SetVersioning(context.Background(), "demo-bucket", types.VersioningConfig{Status: "Enabled"})
	if err != nil {
		t.Fatalf("SetVersioning() error = %v", err)
	}

	cfg, err := service.GetVersioning(context.Background(), "demo-bucket")
	if err != nil {
		t.Fatalf("GetVersioning() error = %v", err)
	}
	if cfg.Status != "Enabled" {
		t.Fatalf("expected Enabled status, got %s", cfg.Status)
	}
}

func TestSetVersioningInvalidStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer server.Close()
	service := createTestService(t, server)

	err := service.SetVersioning(context.Background(), "bucket", types.VersioningConfig{Status: "Unknown"})
	if err == nil {
		t.Fatalf("expected error for invalid status")
	}
}

func TestReplicationNotificationLogging(t *testing.T) {
	tests := []struct {
		name       string
		action     func(Service) error
		expectPath string
	}{
		{
			name: "SetReplication",
			action: func(s Service) error {
				return s.SetReplication(context.Background(), "demo", []byte("<ReplicationConfiguration/>"))
			},
			expectPath: "replication",
		},
		{
			name: "GetReplication",
			action: func(s Service) error {
				_, err := s.GetReplication(context.Background(), "demo")
				return err
			},
			expectPath: "replication",
		},
		{
			name: "DeleteReplication",
			action: func(s Service) error {
				return s.DeleteReplication(context.Background(), "demo")
			},
			expectPath: "replication",
		},
		{
			name: "SetNotification",
			action: func(s Service) error {
				return s.SetNotification(context.Background(), "demo", []byte("<NotificationConfiguration/>"))
			},
			expectPath: "notification",
		},
		{
			name: "GetNotification",
			action: func(s Service) error {
				_, err := s.GetNotification(context.Background(), "demo")
				return err
			},
			expectPath: "notification",
		},
		{
			name: "DeleteNotification",
			action: func(s Service) error {
				return s.DeleteNotification(context.Background(), "demo")
			},
			expectPath: "notification",
		},
		{
			name: "SetLogging",
			action: func(s Service) error {
				return s.SetLogging(context.Background(), "demo", []byte("<BucketLoggingStatus/>"))
			},
			expectPath: "logging",
		},
		{
			name: "GetLogging",
			action: func(s Service) error {
				_, err := s.GetLogging(context.Background(), "demo")
				return err
			},
			expectPath: "logging",
		},
		{
			name: "DeleteLogging",
			action: func(s Service) error {
				return s.DeleteLogging(context.Background(), "demo")
			},
			expectPath: "logging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if _, ok := r.URL.Query()[tt.expectPath]; !ok {
					t.Fatalf("expected %s query flag", tt.expectPath)
				}
				switch r.Method {
				case http.MethodGet:
					_, _ = w.Write([]byte("<ok/>"))
				default:
					w.WriteHeader(http.StatusOK)
				}
			}))
			defer server.Close()

			svc := createTestService(t, server)
			if err := tt.action(svc); err != nil {
				t.Fatalf("%s returned error: %v", tt.name, err)
			}
		})
	}
}
