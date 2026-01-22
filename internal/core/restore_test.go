// Package core internal/core/restore_test.go
package core

import (
	"net/http"
	"testing"
	"time"
)

func TestParseRestoreHeader(t *testing.T) {
	header := http.Header{}
	header.Set("x-amz-restore", "ongoing-request=\"false\", expiry-date=\"Fri, 21 Dec 2012 00:00:00 GMT\"")

	resp := &http.Response{Header: header}
	parser := NewResponseParser()

	info, err := parser.ParseObjectInfo(resp, "bucket", "object")
	if err != nil {
		t.Fatalf("ParseObjectInfo error = %v", err)
	}
	if info.Restore == nil {
		t.Fatalf("expected restore info")
	}
	if info.Restore.OngoingRestore {
		t.Fatalf("expected ongoing restore to be false")
	}
	wantTime, _ := time.Parse(http.TimeFormat, "Fri, 21 Dec 2012 00:00:00 GMT")
	if !info.Restore.ExpiryTime.Equal(wantTime) {
		t.Fatalf("expected expiry time %v, got %v", wantTime, info.Restore.ExpiryTime)
	}
}

func TestParseRestoreHeaderOngoing(t *testing.T) {
	header := http.Header{}
	header.Set("x-amz-restore", "ongoing-request=\"true\"")

	resp := &http.Response{Header: header}
	parser := NewResponseParser()

	info, err := parser.ParseObjectInfo(resp, "bucket", "object")
	if err != nil {
		t.Fatalf("ParseObjectInfo error = %v", err)
	}
	if info.Restore == nil {
		t.Fatalf("expected restore info")
	}
	if !info.Restore.OngoingRestore {
		t.Fatalf("expected ongoing restore to be true")
	}
	if !info.Restore.ExpiryTime.IsZero() {
		t.Fatalf("expected zero expiry time")
	}
}
