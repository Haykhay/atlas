package finding

import "testing"

func valid() Finding {
	return Finding{
		ID:         "ATLAS-TF-001",
		Title:      "S3 bucket without encryption",
		Severity:   SeverityHigh,
		Confidence: 0.92,
		Evidence:   []string{`main.tf:12: resource "aws_s3_bucket" "logs" {}`},
		Origin:     OriginStatic,
	}
}

func TestValidateAcceptsCompleteFinding(t *testing.T) {
	if err := valid().Validate(); err != nil {
		t.Fatalf("expected valid: %v", err)
	}
}

func TestValidateRejectsMissingEvidence(t *testing.T) {
	f := valid()
	f.Evidence = nil
	if err := f.Validate(); err == nil {
		t.Fatal("finding without evidence must fail validation")
	}
}

func TestValidateRejectsBadSeverityConfidenceOrigin(t *testing.T) {
	cases := []func(*Finding){
		func(f *Finding) { f.Severity = "urgent" },
		func(f *Finding) { f.Confidence = 1.5 },
		func(f *Finding) { f.Origin = "guess" },
		func(f *Finding) { f.ID = "" },
		func(f *Finding) { f.Title = "" },
	}
	for i, mutate := range cases {
		f := valid()
		mutate(&f)
		if err := f.Validate(); err == nil {
			t.Errorf("case %d: expected validation error", i)
		}
	}
}
