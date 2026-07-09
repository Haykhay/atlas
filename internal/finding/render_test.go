package finding

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func sample() []Finding {
	low := valid()
	low.ID = "ATLAS-TF-002"
	low.Title = "Missing tags"
	low.Severity = SeverityLow

	crit := valid()
	crit.ID = "ATLAS-TF-003"
	crit.Title = "Public S3 bucket"
	crit.Severity = SeverityCritical
	crit.AffectedResources = []string{"aws_s3_bucket.logs"}
	crit.BusinessImpact = "Data exposure"
	crit.Remediation = "Block public access"
	crit.Pillar = "Security"

	// Deliberately out of order: low first.
	return []Finding{low, crit}
}

func TestRenderJSONRoundTrips(t *testing.T) {
	out := &bytes.Buffer{}
	if err := RenderJSON(out, sample()); err != nil {
		t.Fatalf("render: %v", err)
	}
	var decoded []Finding
	if err := json.Unmarshal(out.Bytes(), &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(decoded) != 2 || decoded[1].ID != "ATLAS-TF-003" {
		t.Fatalf("unexpected decode: %+v", decoded)
	}
}

func TestRenderHumanSortsBySeverityAndShowsDetails(t *testing.T) {
	out := &bytes.Buffer{}
	if err := RenderHuman(out, sample()); err != nil {
		t.Fatalf("render: %v", err)
	}
	s := out.String()

	for _, want := range []string{"[CRITICAL]", "Public S3 bucket", "92%", "aws_s3_bucket.logs", "main.tf:12", "Data exposure", "Block public access", "Security"} {
		if !strings.Contains(s, want) {
			t.Errorf("output missing %q:\n%s", want, s)
		}
	}
	if strings.Index(s, "[CRITICAL]") > strings.Index(s, "[LOW]") {
		t.Errorf("critical must render before low:\n%s", s)
	}
}
