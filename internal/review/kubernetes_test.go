package review

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Haykhay/atlas/internal/finding"
)

const fixtureManifest = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 1
  template:
    spec:
      containers:
        - name: app
          image: nginx:latest
`

func manifestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "app.yaml"), []byte(fixtureManifest), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return dir
}

func TestKubernetesRunStaticOnly(t *testing.T) {
	k := &Kubernetes{}
	report, err := k.Run(context.Background(), manifestDir(t))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !hasOrigin(report.Findings, finding.OriginStatic) {
		t.Fatalf("expected static findings: %+v", report.Findings)
	}
	if len(report.PillarScores) != 6 {
		t.Fatalf("expected 6 pillar scores, got %d", len(report.PillarScores))
	}
}

func TestKubernetesRunMergesAIFindings(t *testing.T) {
	ai := &stubAI{text: `[
  {"id":"ATLAS-AI-001","title":"No network policy","severity":"medium","confidence":0.7,
   "evidence":["app.yaml:1: kind: Deployment"],"affected_resources":["Deployment/web"],
   "business_impact":"i","remediation":"r","pillar":"Security"}
]`}
	k := &Kubernetes{AI: ai}
	report, err := k.Run(context.Background(), manifestDir(t))
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !hasOrigin(report.Findings, finding.OriginAI) {
		t.Fatalf("expected ai findings: %+v", report.Findings)
	}
}

func TestKubernetesRunDegradesGracefully(t *testing.T) {
	k := &Kubernetes{AI: &stubAI{err: errors.New("provider down")}}
	report, err := k.Run(context.Background(), manifestDir(t))
	if err != nil {
		t.Fatalf("run must not fail on AI error: %v", err)
	}
	if len(report.Warnings) == 0 || !hasOrigin(report.Findings, finding.OriginStatic) {
		t.Fatalf("expected warning + static findings: %+v", report)
	}
}
