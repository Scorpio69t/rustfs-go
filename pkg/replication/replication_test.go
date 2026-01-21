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
