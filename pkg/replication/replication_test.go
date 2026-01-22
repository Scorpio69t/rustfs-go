package replication

import (
	"strings"
	"testing"
)

func TestReplicationConfigXML(t *testing.T) {
	cfg := ReplicationConfig{
		Role: "arn:aws:iam::123456789012:role/replication",
		Rules: []Rule{
			{
				ID:       "rule-1",
				Status:   Enabled,
				Priority: 1,
				Filter: Filter{
					Prefix: "logs/",
				},
				Destination: Destination{
					Bucket: "arn:aws:s3:::dest-bucket",
				},
			},
		},
	}

	data, err := cfg.ToXML()
	if err != nil {
		t.Fatalf("ToXML() error = %v", err)
	}
	if !strings.Contains(string(data), "<ReplicationConfiguration") {
		t.Fatalf("expected ReplicationConfiguration element")
	}

	parsed, err := ParseConfig(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if len(parsed.Rules) != 1 || parsed.Rules[0].Destination.Bucket == "" {
		t.Fatalf("unexpected parsed config: %+v", parsed)
	}
}

func TestReplicationConfigNormalizeErrors(t *testing.T) {
	cfg := ReplicationConfig{
		Rules: []Rule{
			{
				Status:      Status("Invalid"),
				Destination: Destination{Bucket: "arn:aws:s3:::dest-bucket"},
			},
		},
	}

	if err := cfg.Normalize(); err == nil {
		t.Fatalf("expected invalid status error")
	}

	cfg = ReplicationConfig{
		Rules: []Rule{
			{
				Status:      Enabled,
				Destination: Destination{},
			},
		},
	}

	if err := cfg.Normalize(); err == nil {
		t.Fatalf("expected missing destination bucket error")
	}
}
