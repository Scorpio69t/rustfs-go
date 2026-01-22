package policy

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func TestPostPolicyBase64(t *testing.T) {
	p := NewPostPolicy()
	expires := time.Date(2026, 1, 21, 0, 0, 0, 0, time.UTC)
	if err := p.SetExpires(expires); err != nil {
		t.Fatalf("SetExpires() error = %v", err)
	}
	if err := p.SetBucket("test-bucket"); err != nil {
		t.Fatalf("SetBucket() error = %v", err)
	}
	if err := p.SetKey("test-key"); err != nil {
		t.Fatalf("SetKey() error = %v", err)
	}
	if err := p.SetContentType("text/plain"); err != nil {
		t.Fatalf("SetContentType() error = %v", err)
	}
	if err := p.SetContentLengthRange(1, 10); err != nil {
		t.Fatalf("SetContentLengthRange() error = %v", err)
	}

	encoded, err := p.Base64()
	if err != nil {
		t.Fatalf("Base64() error = %v", err)
	}

	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("DecodeString() error = %v", err)
	}

	var doc struct {
		Expiration string        `json:"expiration"`
		Conditions []interface{} `json:"conditions"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if doc.Expiration != "2026-01-21T00:00:00.000Z" {
		t.Fatalf("expiration = %q, want %q", doc.Expiration, "2026-01-21T00:00:00.000Z")
	}
	if !hasCondition(doc.Conditions, "eq", "$bucket", "test-bucket") {
		t.Fatalf("missing bucket condition in %v", doc.Conditions)
	}
	if !hasCondition(doc.Conditions, "eq", "$key", "test-key") {
		t.Fatalf("missing key condition in %v", doc.Conditions)
	}
	if !hasCondition(doc.Conditions, "eq", "$Content-Type", "text/plain") {
		t.Fatalf("missing content-type condition in %v", doc.Conditions)
	}
	if !hasContentLengthRange(doc.Conditions, 1, 10) {
		t.Fatalf("missing content-length-range condition in %v", doc.Conditions)
	}
}

func TestPostPolicyBase64RequiresExpiration(t *testing.T) {
	p := NewPostPolicy()
	if err := p.SetBucket("test-bucket"); err != nil {
		t.Fatalf("SetBucket() error = %v", err)
	}
	if err := p.SetKey("test-key"); err != nil {
		t.Fatalf("SetKey() error = %v", err)
	}
	if _, err := p.Base64(); err == nil {
		t.Fatal("Base64() expected error when expiration is missing")
	}
}

func hasCondition(conditions []interface{}, matchType, condition, value string) bool {
	for _, c := range conditions {
		arr, ok := c.([]interface{})
		if !ok || len(arr) != 3 {
			continue
		}
		if arr[0] == matchType && arr[1] == condition && arr[2] == value {
			return true
		}
	}
	return false
}

func hasContentLengthRange(conditions []interface{}, min, max int64) bool {
	for _, c := range conditions {
		arr, ok := c.([]interface{})
		if !ok || len(arr) != 3 {
			continue
		}
		if arr[0] != "content-length-range" {
			continue
		}
		minVal, okMin := arr[1].(float64)
		maxVal, okMax := arr[2].(float64)
		if okMin && okMax && int64(minVal) == min && int64(maxVal) == max {
			return true
		}
	}
	return false
}
