package sse

import (
	"encoding/base64"
	"net/http"
	"testing"
)

func TestNewSSES3(t *testing.T) {
	enc := NewSSES3()
	if enc == nil {
		t.Fatal("NewSSES3 returned nil")
	}
	if enc.Type() != SSES3 {
		t.Errorf("Expected type SSES3, got %v", enc.Type())
	}
}

func TestS3ApplyHeaders(t *testing.T) {
	enc := NewSSES3()
	headers := make(http.Header)
	enc.ApplyHeaders(headers)

	if got := headers.Get("X-Amz-Server-Side-Encryption"); got != "AES256" {
		t.Errorf("Expected AES256, got %s", got)
	}
}

func TestNewSSEC(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		wantErr bool
	}{
		{
			name:    "valid 32-byte key",
			key:     make([]byte, 32),
			wantErr: false,
		},
		{
			name:    "invalid 16-byte key",
			key:     make([]byte, 16),
			wantErr: true,
		},
		{
			name:    "invalid 64-byte key",
			key:     make([]byte, 64),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc, err := NewSSEC(tt.key)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				if err != ErrInvalidKeySize {
					t.Errorf("Expected ErrInvalidKeySize, got %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if enc == nil {
					t.Fatal("NewSSEC returned nil")
				}
			}
		})
	}
}

func TestCApplyHeaders(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	enc, err := NewSSEC(key)
	if err != nil {
		t.Fatalf("Failed to create SSE-C: %v", err)
	}

	headers := make(http.Header)
	enc.ApplyHeaders(headers)

	// Check algorithm
	if got := headers.Get("X-Amz-Server-Side-Encryption-Customer-Algorithm"); got != "AES256" {
		t.Errorf("Expected AES256, got %s", got)
	}

	// Check key
	expectedKey := base64.StdEncoding.EncodeToString(key)
	if got := headers.Get("X-Amz-Server-Side-Encryption-Customer-Key"); got != expectedKey {
		t.Errorf("Key mismatch")
	}

	// Check key MD5
	if got := headers.Get("X-Amz-Server-Side-Encryption-Customer-Key-MD5"); got == "" {
		t.Error("Expected MD5 header, got empty")
	}
}

func TestNewSSEKMS(t *testing.T) {
	keyID := "arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012"
	context := map[string]string{
		"Department": "Finance",
	}

	enc := NewSSEKMS(keyID, context)
	if enc == nil {
		t.Fatal("NewSSEKMS returned nil")
	}
	if enc.KeyID != keyID {
		t.Errorf("Expected keyID %s, got %s", keyID, enc.KeyID)
	}
	if enc.Type() != SSEKMS {
		t.Errorf("Expected type SSEKMS, got %v", enc.Type())
	}
}

func TestKMSApplyHeaders(t *testing.T) {
	keyID := "test-key-id"
	context := map[string]string{
		"Project": "TestProject",
	}

	enc := NewSSEKMS(keyID, context)
	headers := make(http.Header)
	enc.ApplyHeaders(headers)

	// Check encryption type
	if got := headers.Get("X-Amz-Server-Side-Encryption"); got != "aws:kms" {
		t.Errorf("Expected aws:kms, got %s", got)
	}

	// Check key ID
	if got := headers.Get("X-Amz-Server-Side-Encryption-Aws-Kms-Key-Id"); got != keyID {
		t.Errorf("Expected %s, got %s", keyID, got)
	}

	// Check context exists
	if got := headers.Get("X-Amz-Server-Side-Encryption-Context"); got == "" {
		t.Error("Expected encryption context, got empty")
	}
}

func TestNewConfiguration(t *testing.T) {
	config := NewConfiguration()
	if config == nil {
		t.Fatal("NewConfiguration returned nil")
	}
	if len(config.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(config.Rules))
	}
	if config.Rules[0].ApplySSEByDefault.SSEAlgorithm != "AES256" {
		t.Errorf("Expected AES256, got %s", config.Rules[0].ApplySSEByDefault.SSEAlgorithm)
	}
}

func TestNewKMSConfiguration(t *testing.T) {
	keyID := "test-key-id"
	config := NewKMSConfiguration(keyID)
	if config == nil {
		t.Fatal("NewKMSConfiguration returned nil")
	}
	if len(config.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(config.Rules))
	}
	rule := config.Rules[0]
	if rule.ApplySSEByDefault.SSEAlgorithm != "aws:kms" {
		t.Errorf("Expected aws:kms, got %s", rule.ApplySSEByDefault.SSEAlgorithm)
	}
	if rule.ApplySSEByDefault.KMSMasterKeyID != keyID {
		t.Errorf("Expected %s, got %s", keyID, rule.ApplySSEByDefault.KMSMasterKeyID)
	}
	if !rule.BucketKeyEnabled {
		t.Error("Expected BucketKeyEnabled to be true")
	}
}
