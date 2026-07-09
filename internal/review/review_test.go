package review

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Haykhay/atlas/internal/finding"
	"github.com/Haykhay/atlas/internal/provider"
)

const fixtureTF = `resource "aws_db_instance" "app" {
  storage_encrypted = false
}
`

func fixtureDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(fixtureTF), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return dir
}

type stubAI struct {
	text string
	err  error
}

func (s *stubAI) Name() string { return "stub" }
func (s *stubAI) Complete(_ context.Context, _ provider.Request) (provider.Response, error) {
	if s.err != nil {
		return provider.Response{}, s.err
	}
	return provider.Response{Text: s.text, Model: "stub"}, nil
}
func (s *stubAI) Status(_ context.Context) error { return nil }

func hasOrigin(fs []finding.Finding, o finding.Origin) bool {
	for _, f := range fs {
		if f.Origin == o {
			return true
		}
	}
	return false
}

func TestRunStaticOnly(t *testing.T) {
	r := &Terraform{}
	report, err := r.Run(context.Background(), fixtureDir(t))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !hasOrigin(report.Findings, finding.OriginStatic) {
		t.Fatalf("expected static findings: %+v", report.Findings)
	}
	if len(report.PillarScores) != 6 {
		t.Fatalf("expected 6 pillar scores, got %d", len(report.PillarScores))
	}
	if len(report.Warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", report.Warnings)
	}
}

func TestRunMergesAIFindings(t *testing.T) {
	ai := &stubAI{text: `[
  {"id":"ATLAS-AI-001","title":"Broad IAM","severity":"high","confidence":0.8,
   "evidence":["main.tf:1: x"],"affected_resources":["aws_iam_policy.p"],
   "business_impact":"i","remediation":"r","pillar":"Security"}
]`}
	r := &Terraform{AI: ai}
	report, err := r.Run(context.Background(), fixtureDir(t))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !hasOrigin(report.Findings, finding.OriginAI) || !hasOrigin(report.Findings, finding.OriginStatic) {
		t.Fatalf("expected both origins: %+v", report.Findings)
	}
}

func TestRunDegradesGracefullyOnAIError(t *testing.T) {
	r := &Terraform{AI: &stubAI{err: errors.New("provider down")}}
	report, err := r.Run(context.Background(), fixtureDir(t))
	if err != nil {
		t.Fatalf("run must not fail on AI error: %v", err)
	}
	if !hasOrigin(report.Findings, finding.OriginStatic) {
		t.Fatal("static findings must survive AI failure")
	}
	if len(report.Warnings) == 0 {
		t.Fatal("expected a warning about AI failure")
	}
}
