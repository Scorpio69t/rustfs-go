package objectlock

import "testing"

func TestConfigNormalizeDefaults(t *testing.T) {
	cfg := Config{}
	if err := cfg.Normalize(); err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if cfg.ObjectLockEnabled != ObjectLockEnabledValue {
		t.Fatalf("expected %s, got %s", ObjectLockEnabledValue, cfg.ObjectLockEnabled)
	}
}

func TestConfigNormalizeInvalidState(t *testing.T) {
	cfg := Config{ObjectLockEnabled: "Disabled"}
	if err := cfg.Normalize(); err != ErrInvalidObjectLockState {
		t.Fatalf("expected ErrInvalidObjectLockState, got %v", err)
	}
}

func TestConfigNormalizeRetentionValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr error
	}{
		{
			name: "invalid retention mode",
			cfg: Config{
				Rule: &Rule{
					DefaultRetention: DefaultRetention{
						Mode: "INVALID",
						Days: 1,
					},
				},
			},
			wantErr: ErrInvalidRetentionMode,
		},
		{
			name: "missing retention period",
			cfg: Config{
				Rule: &Rule{
					DefaultRetention: DefaultRetention{
						Mode: RetentionGovernance,
					},
				},
			},
			wantErr: ErrInvalidRetentionPeriod,
		},
		{
			name: "both days and years",
			cfg: Config{
				Rule: &Rule{
					DefaultRetention: DefaultRetention{
						Mode:  RetentionCompliance,
						Days:  1,
						Years: 1,
					},
				},
			},
			wantErr: ErrInvalidRetentionPeriod,
		},
		{
			name: "valid retention days",
			cfg: Config{
				Rule: &Rule{
					DefaultRetention: DefaultRetention{
						Mode: RetentionGovernance,
						Days: 7,
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.cfg.Normalize(); err != tt.wantErr {
				t.Fatalf("Normalize() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestLegalHoldStatusValid(t *testing.T) {
	if !LegalHoldOn.IsValid() {
		t.Fatalf("expected LegalHoldOn to be valid")
	}
	if LegalHoldStatus("INVALID").IsValid() {
		t.Fatalf("expected invalid legal hold status")
	}
}
