package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Haykhay/atlas/internal/credentials"
	"github.com/Haykhay/atlas/internal/review"
)

const insecureFixture = `resource "aws_security_group" "web" {
  ingress {
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_instance" "app" {
  storage_encrypted = false
}
`

func reviewFixtureDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(insecureFixture), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	return dir
}

func TestReviewTerraformOfflineHuman(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	out, err := runCmd(t, "", "review", "terraform", reviewFixtureDir(t), "--offline")
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	for _, want := range []string{"AWS Well-Architected Scores", "Security", "[CRITICAL]", "[HIGH]", "aws_security_group.web"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestReviewTerraformJSONFormat(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	out, err := runCmd(t, "", "review", "terraform", reviewFixtureDir(t), "--offline", "--format", "json")
	if err != nil {
		t.Fatalf("review: %v", err)
	}
	var report review.Report
	if err := json.Unmarshal([]byte(out), &report); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, out)
	}
	if len(report.PillarScores) != 6 {
		t.Fatalf("expected 6 pillar scores, got %d", len(report.PillarScores))
	}
	if len(report.Findings) == 0 || report.Findings[0].Origin != "static" {
		t.Fatalf("expected static findings: %+v", report.Findings)
	}
}
