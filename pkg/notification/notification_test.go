package notification

import (
	"strings"
	"testing"
)

func TestConfigurationXML(t *testing.T) {
	cfg := Configuration{
		QueueConfigs: []QueueConfig{
			{
				Config: Config{
					ID:     "queue-1",
					Events: []EventType{ObjectCreatedAll},
				},
				Queue: "arn:aws:sqs:us-east-1:123456789012:queue1",
			},
		},
	}

	data, err := cfg.ToXML()
	if err != nil {
		t.Fatalf("ToXML() error = %v", err)
	}
	if !strings.Contains(string(data), "<NotificationConfiguration") {
		t.Fatalf("expected NotificationConfiguration element")
	}

	parsed, err := ParseConfig(strings.NewReader(string(data)))
	if err != nil {
		t.Fatalf("ParseConfig() error = %v", err)
	}
	if len(parsed.QueueConfigs) != 1 || parsed.QueueConfigs[0].Queue == "" {
		t.Fatalf("unexpected parsed config: %+v", parsed)
	}
}
