package restore

import (
	"encoding/xml"
	"strings"
	"testing"

	s3select "github.com/Scorpio69t/rustfs-go/pkg/select"
)

func TestRestoreRequestNormalizeDefaults(t *testing.T) {
	req := &RestoreRequest{}
	req.Normalize()

	if req.XMLNS != defaultXMLNS {
		t.Fatalf("expected XMLNS %q, got %q", defaultXMLNS, req.XMLNS)
	}
}

func TestRestoreRequestSetters(t *testing.T) {
	req := &RestoreRequest{}

	req.SetDays(7)
	req.SetType(RestoreSelect)
	req.SetTier(TierExpedited)
	req.SetDescription("restore test")
	req.SetGlacierJobParameters(GlacierJobParameters{Tier: TierStandard})

	if req.Days == nil || *req.Days != 7 {
		t.Fatalf("expected Days to be set to 7, got %#v", req.Days)
	}
	if req.Type == nil || *req.Type != RestoreSelect {
		t.Fatalf("expected Type to be set to %q", RestoreSelect)
	}
	if req.Tier == nil || *req.Tier != TierExpedited {
		t.Fatalf("expected Tier to be set to %q", TierExpedited)
	}
	if req.Description == nil || *req.Description != "restore test" {
		t.Fatalf("expected Description to be set")
	}
	if req.GlacierJobParameters == nil || req.GlacierJobParameters.Tier != TierStandard {
		t.Fatalf("expected GlacierJobParameters to be set")
	}
}

func TestRestoreRequestXMLMarshaling(t *testing.T) {
	req := RestoreRequest{}
	req.SetType(RestoreSelect)
	req.SetDays(1)
	req.SetSelectParameters(SelectParameters{
		ExpressionType: s3select.QueryExpressionTypeSQL,
		Expression:     "SELECT * FROM S3Object",
		InputSerialization: s3select.InputSerialization{
			CSV: &s3select.CSVInputOptions{},
		},
		OutputSerialization: s3select.OutputSerialization{
			CSV: &s3select.CSVOutputOptions{},
		},
	})
	req.Normalize()

	data, err := xml.Marshal(req)
	if err != nil {
		t.Fatalf("xml.Marshal error = %v", err)
	}

	encoded := string(data)
	if !strings.Contains(encoded, "RestoreRequest") {
		t.Fatalf("expected RestoreRequest element, got %s", encoded)
	}
	if !strings.Contains(encoded, `xmlns="`+defaultXMLNS+`"`) {
		t.Fatalf("expected xmlns attribute, got %s", encoded)
	}
	if !strings.Contains(encoded, "SelectParameters") {
		t.Fatalf("expected SelectParameters element, got %s", encoded)
	}
}
