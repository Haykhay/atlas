package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Haykhay/atlas/internal/credentials"
)

func TestFixTerraformNothingToFix(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	dir := t.TempDir()
	clean := "resource \"aws_cloudwatch_log_group\" \"app\" {\n  name              = \"app\"\n  retention_in_days = 30\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "main.tf"), []byte(clean), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	out, err := runCmd(t, "", "fix", "terraform", dir)
	if err != nil {
		t.Fatalf("fix must succeed with no findings and no provider: %v", err)
	}
	if !strings.Contains(out, "No findings to fix.") {
		t.Fatalf("unexpected output: %s", out)
	}
}

func TestFixTerraformRequiresProviderWhenFindingsExist(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	_, err := runCmd(t, "", "fix", "terraform", reviewFixtureDir(t))
	if err == nil || !strings.Contains(err.Error(), "atlas configure") {
		t.Fatalf("fix without provider must point to configure: %v", err)
	}
}
