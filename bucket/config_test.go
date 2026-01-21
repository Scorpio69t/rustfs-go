// Package bucket bucket/config_test.go
package bucket

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Scorpio69t/rustfs-go/pkg/acl"
	"github.com/Scorpio69t/rustfs-go/pkg/cors"
	"github.com/Scorpio69t/rustfs-go/pkg/notification"
	"github.com/Scorpio69t/rustfs-go/pkg/objectlock"
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

func TestSetAndGetObjectLockConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["object-lock"]; !ok {
			t.Fatalf("expected object-lock query flag")
		}
		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "<ObjectLockEnabled>Enabled</ObjectLockEnabled>") {
				t.Fatalf("expected object lock enabled in body, got %s", string(body))
			}
			if !strings.Contains(string(body), "<Mode>GOVERNANCE</Mode>") {
				t.Fatalf("expected retention mode in body, got %s", string(body))
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<ObjectLockConfiguration><ObjectLockEnabled>Enabled</ObjectLockEnabled><Rule><DefaultRetention><Mode>GOVERNANCE</Mode><Days>1</Days></DefaultRetention></Rule></ObjectLockConfiguration>`))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	service := createTestService(t, server)
	config := objectlock.Config{
		Rule: &objectlock.Rule{
			DefaultRetention: objectlock.DefaultRetention{
				Mode: objectlock.RetentionGovernance,
				Days: 1,
			},
		},
	}

	if err := service.SetObjectLockConfig(context.Background(), "demo-bucket", config); err != nil {
		t.Fatalf("SetObjectLockConfig() error = %v", err)
	}

	got, err := service.GetObjectLockConfig(context.Background(), "demo-bucket")
	if err != nil {
		t.Fatalf("GetObjectLockConfig() error = %v", err)
	}
	if got.ObjectLockEnabled != objectlock.ObjectLockEnabledValue {
		t.Fatalf("expected enabled, got %s", got.ObjectLockEnabled)
	}
	if got.Rule == nil || got.Rule.DefaultRetention.Mode != objectlock.RetentionGovernance {
		t.Fatalf("expected governance retention mode")
	}
}

func TestSetGetDeleteCORS(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<CORSConfiguration>
  <CORSRule>
    <AllowedOrigin>*</AllowedOrigin>
    <AllowedMethod>GET</AllowedMethod>
  </CORSRule>
</CORSConfiguration>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["cors"]; !ok {
			t.Fatalf("expected cors query flag")
		}
		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "<CORSConfiguration") {
				t.Fatalf("expected cors config in body, got %s", string(body))
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(responseXML))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	service := createTestService(t, server)
	config := cors.NewConfig([]cors.Rule{
		{
			AllowedOrigin: []string{"*"},
			AllowedMethod: []string{"GET"},
		},
	})

	if err := service.SetCORS(context.Background(), "demo-bucket", config); err != nil {
		t.Fatalf("SetCORS() error = %v", err)
	}

	got, err := service.GetCORS(context.Background(), "demo-bucket")
	if err != nil {
		t.Fatalf("GetCORS() error = %v", err)
	}
	if len(got.CORSRules) != 1 || got.CORSRules[0].AllowedOrigin[0] != "*" {
		t.Fatalf("unexpected CORS config: %+v", got)
	}

	if err := service.DeleteCORS(context.Background(), "demo-bucket"); err != nil {
		t.Fatalf("DeleteCORS() error = %v", err)
	}
}

func TestBucketTaggingCRUD(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<Tagging>
  <TagSet>
    <Tag>
      <Key>env</Key>
      <Value>prod</Value>
    </Tag>
    <Tag>
      <Key>team</Key>
      <Value>storage</Value>
    </Tag>
  </TagSet>
</Tagging>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["tagging"]; !ok {
			t.Fatalf("expected tagging query flag")
		}
		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "<Tagging>") {
				t.Fatalf("expected tagging XML in body, got %s", string(body))
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(responseXML))
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	service := createTestService(t, server)
	tags := map[string]string{"env": "prod", "team": "storage"}

	if err := service.SetTagging(context.Background(), "demo-bucket", tags); err != nil {
		t.Fatalf("SetTagging() error = %v", err)
	}

	got, err := service.GetTagging(context.Background(), "demo-bucket")
	if err != nil {
		t.Fatalf("GetTagging() error = %v", err)
	}
	if got["env"] != "prod" || got["team"] != "storage" {
		t.Fatalf("unexpected tags: %+v", got)
	}

	if err := service.DeleteTagging(context.Background(), "demo-bucket"); err != nil {
		t.Fatalf("DeleteTagging() error = %v", err)
	}
}

func TestBucketACLSetGet(t *testing.T) {
	responseXML := `<?xml version="1.0" encoding="UTF-8"?>
<AccessControlPolicy>
  <Owner>
    <ID>owner-id</ID>
    <DisplayName>owner</DisplayName>
  </Owner>
  <AccessControlList>
    <Grant>
      <Grantee xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="CanonicalUser">
        <ID>grantee-id</ID>
      </Grantee>
      <Permission>READ</Permission>
    </Grant>
  </AccessControlList>
</AccessControlPolicy>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["acl"]; !ok {
			t.Fatalf("expected acl query flag")
		}
		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "<AccessControlPolicy") {
				t.Fatalf("expected ACL XML in body, got %s", string(body))
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(responseXML))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	service := createTestService(t, server)
	policy := acl.ACL{
		Owner: acl.Owner{ID: "owner-id", DisplayName: "owner"},
		Grants: []acl.Grant{
			{
				Grantee:    acl.Grantee{Type: "CanonicalUser", ID: "grantee-id"},
				Permission: acl.PermissionRead,
			},
		},
	}

	if err := service.SetACL(context.Background(), "demo-bucket", policy); err != nil {
		t.Fatalf("SetACL() error = %v", err)
	}

	got, err := service.GetACL(context.Background(), "demo-bucket")
	if err != nil {
		t.Fatalf("GetACL() error = %v", err)
	}
	if len(got.Grants) != 1 || got.Grants[0].Permission != acl.PermissionRead {
		t.Fatalf("unexpected ACL grants: %+v", got.Grants)
	}
}

func TestGetReplicationMetrics(t *testing.T) {
	responseJSON := `{"replicationCount":5,"completedReplicationSize":2048}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["replication-metrics"]; !ok {
			t.Fatalf("expected replication-metrics query flag")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	service := createTestService(t, server)
	metrics, err := service.GetReplicationMetrics(context.Background(), "demo-bucket")
	if err != nil {
		t.Fatalf("GetReplicationMetrics() error = %v", err)
	}
	if metrics.ReplicatedCount != 5 || metrics.ReplicatedSize != 2048 {
		t.Fatalf("unexpected metrics: %+v", metrics)
	}
}

func TestListenNotification(t *testing.T) {
	responseLine := `{"Records":[{"eventName":"s3:ObjectCreated:Put"}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["ping"]; !ok {
			t.Fatalf("expected ping query flag")
		}
		events := r.URL.Query()["events"]
		if len(events) != 1 || events[0] != string(notification.ObjectCreatedAll) {
			t.Fatalf("unexpected events query: %v", events)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseLine + "\n"))
	}))
	defer server.Close()

	service := createTestService(t, server)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ch := service.ListenNotification(ctx, "demo-bucket", "", "", []notification.EventType{notification.ObjectCreatedAll})
	select {
	case info := <-ch:
		if info.Err != nil {
			t.Fatalf("ListenNotification() error = %v", info.Err)
		}
		if len(info.Records) != 1 || info.Records[0].EventName != "s3:ObjectCreated:Put" {
			t.Fatalf("unexpected notification info: %+v", info)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for notification")
	}
}
