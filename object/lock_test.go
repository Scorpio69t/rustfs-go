// Package object object/lock_test.go
package object

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Scorpio69t/rustfs-go/pkg/objectlock"
)

func TestSetLegalHold(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if _, ok := r.URL.Query()["legal-hold"]; !ok {
			t.Fatalf("expected legal-hold query flag")
		}
		if r.URL.Query().Get("versionId") != "ver-1" {
			t.Fatalf("expected versionId query")
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "<Status>ON</Status>") {
			t.Fatalf("expected legal hold status in body, got %s", string(body))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	err := service.SetLegalHold(context.Background(), "bucket", "object", objectlock.LegalHoldOn, WithLegalHoldVersionID("ver-1"))
	if err != nil {
		t.Fatalf("SetLegalHold() error = %v", err)
	}
}

func TestGetLegalHold(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["legal-hold"]; !ok {
			t.Fatalf("expected legal-hold query flag")
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<LegalHold><Status>OFF</Status></LegalHold>`))
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	status, err := service.GetLegalHold(context.Background(), "bucket", "object")
	if err != nil {
		t.Fatalf("GetLegalHold() error = %v", err)
	}
	if status != objectlock.LegalHoldOff {
		t.Fatalf("expected OFF, got %s", status)
	}
}

func TestSetRetention(t *testing.T) {
	retainUntil := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	expectedTime := retainUntil.Format(time.RFC3339)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if _, ok := r.URL.Query()["retention"]; !ok {
			t.Fatalf("expected retention query flag")
		}
		if r.URL.Query().Get("versionId") != "ver-2" {
			t.Fatalf("expected versionId query")
		}
		if r.Header.Get("x-amz-bypass-governance-retention") != "true" {
			t.Fatalf("expected governance bypass header")
		}
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "<Mode>GOVERNANCE</Mode>") {
			t.Fatalf("expected retention mode in body, got %s", bodyStr)
		}
		if !strings.Contains(bodyStr, expectedTime) {
			t.Fatalf("expected retain-until date in body, got %s", bodyStr)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	err := service.SetRetention(
		context.Background(),
		"bucket",
		"object",
		objectlock.RetentionGovernance,
		retainUntil,
		WithRetentionVersionID("ver-2"),
		WithGovernanceBypass(),
	)
	if err != nil {
		t.Fatalf("SetRetention() error = %v", err)
	}
}

func TestGetRetention(t *testing.T) {
	retainUntil := time.Date(2027, time.February, 3, 4, 5, 6, 0, time.UTC).Format(time.RFC3339)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.URL.Query()["retention"]; !ok {
			t.Fatalf("expected retention query flag")
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<Retention><Mode>COMPLIANCE</Mode><RetainUntilDate>` + retainUntil + `</RetainUntilDate></Retention>`))
	}))
	defer server.Close()

	service := createAdvancedTestService(t, server)
	mode, until, err := service.GetRetention(context.Background(), "bucket", "object")
	if err != nil {
		t.Fatalf("GetRetention() error = %v", err)
	}
	if mode != objectlock.RetentionCompliance {
		t.Fatalf("expected COMPLIANCE, got %s", mode)
	}
	if until.Format(time.RFC3339) != retainUntil {
		t.Fatalf("unexpected retain-until date: %s", until.Format(time.RFC3339))
	}
}
