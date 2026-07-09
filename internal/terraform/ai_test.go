package terraform

import (
	"strings"
	"testing"

	"github.com/Haykhay/atlas/internal/finding"
)

func TestFilesPromptListsFilesSorted(t *testing.T) {
	p := FilesPrompt(map[string]string{
		"vpc.tf":  `resource "aws_vpc" "main" {}`,
		"main.tf": `resource "aws_s3_bucket" "logs" {}`,
	})
	if strings.Index(p, "main.tf") > strings.Index(p, "vpc.tf") {
		t.Fatalf("files must be sorted: %s", p)
	}
	if !strings.Contains(p, "aws_s3_bucket") {
		t.Fatalf("content missing: %s", p)
	}
}

func TestParseAIFindingsExtractsAndValidates(t *testing.T) {
	resp := `Here is my analysis:
[
  {
    "id": "ATLAS-AI-001",
    "title": "IAM policy overly broad",
    "severity": "high",
    "confidence": 92,
    "evidence": ["iam.tf:4: Action = \"*\""],
    "affected_resources": ["aws_iam_policy.admin"],
    "business_impact": "Full account compromise blast radius",
    "remediation": "Scope actions to required services",
    "pillar": "Security"
  },
  {
    "id": "ATLAS-AI-002",
    "title": "No evidence here",
    "severity": "low",
    "confidence": 0.5,
    "evidence": []
  }
]
Hope this helps!`

	fs, err := ParseAIFindings(resp)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(fs) != 1 {
		t.Fatalf("expected the invalid finding dropped, got %d: %+v", len(fs), fs)
	}
	f := fs[0]
	if f.Origin != finding.OriginAI {
		t.Errorf("origin must be forced to ai, got %q", f.Origin)
	}
	if f.Confidence != 0.92 {
		t.Errorf("confidence 92 must normalize to 0.92, got %v", f.Confidence)
	}
	if err := f.Validate(); err != nil {
		t.Errorf("surviving finding must validate: %v", err)
	}
}

func TestParseAIFindingsErrorsWithoutArray(t *testing.T) {
	if _, err := ParseAIFindings("I could not find any issues."); err == nil {
		t.Fatal("expected error when no JSON array present")
	}
}
