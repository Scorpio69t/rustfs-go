package types

import "testing"

func TestChecksumTypeString(t *testing.T) {
	cases := []struct {
		value    ChecksumType
		expected string
	}{
		{ChecksumNone, ""},
		{ChecksumCRC32, "CRC32"},
		{ChecksumCRC32C, "CRC32C"},
		{ChecksumSHA1, "SHA1"},
		{ChecksumSHA256, "SHA256"},
		{ChecksumCRC64NVME, "CRC64NVME"},
		{ChecksumType(999), ""},
	}

	for _, tc := range cases {
		if got := tc.value.String(); got != tc.expected {
			t.Fatalf("ChecksumType(%d).String() = %q, want %q", tc.value, got, tc.expected)
		}
	}
}

func TestRetentionModeIsValid(t *testing.T) {
	if !RetentionGovernance.IsValid() {
		t.Fatalf("RetentionGovernance should be valid")
	}
	if !RetentionCompliance.IsValid() {
		t.Fatalf("RetentionCompliance should be valid")
	}
	if RetentionMode("INVALID").IsValid() {
		t.Fatalf("expected invalid retention mode to be rejected")
	}
}

func TestLegalHoldStatusIsValid(t *testing.T) {
	if !LegalHoldOn.IsValid() {
		t.Fatalf("LegalHoldOn should be valid")
	}
	if !LegalHoldOff.IsValid() {
		t.Fatalf("LegalHoldOff should be valid")
	}
	if LegalHoldStatus("INVALID").IsValid() {
		t.Fatalf("expected invalid legal hold status to be rejected")
	}
}
