package cors

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseBucketCORSConfig(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<CORSConfiguration>
  <CORSRule>
    <AllowedOrigin>https://example.com</AllowedOrigin>
    <AllowedMethod>get</AllowedMethod>
    <AllowedMethod>put</AllowedMethod>
  </CORSRule>
</CORSConfiguration>`

	cfg, err := ParseBucketCORSConfig(strings.NewReader(xmlData))
	if err != nil {
		t.Fatalf("ParseBucketCORSConfig() error = %v", err)
	}
	if len(cfg.CORSRules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(cfg.CORSRules))
	}
	if cfg.CORSRules[0].AllowedMethod[0] != "GET" || cfg.CORSRules[0].AllowedMethod[1] != "PUT" {
		t.Fatalf("expected methods to be uppercased, got %+v", cfg.CORSRules[0].AllowedMethod)
	}
	if cfg.XMLNS == "" {
		t.Fatalf("expected XMLNS to be set")
	}
}

func TestToXML(t *testing.T) {
	cfg := NewConfig([]Rule{
		{
			AllowedOrigin: []string{"*"},
			AllowedMethod: []string{"GET"},
		},
	})
	data, err := cfg.ToXML()
	if err != nil {
		t.Fatalf("ToXML() error = %v", err)
	}
	if !bytes.HasPrefix(data, []byte("<?xml")) {
		t.Fatalf("expected XML header")
	}
	if !strings.Contains(string(data), "<CORSConfiguration") {
		t.Fatalf("expected CORSConfiguration element")
	}
}
