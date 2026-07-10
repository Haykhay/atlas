package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Haykhay/atlas/internal/credentials"
)

func TestDocumentTerraformOfflineProducesSkeleton(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	out, err := runCmd(t, "", "document", "terraform", reviewFixtureDir(t))
	if err != nil {
		t.Fatalf("document: %v", err)
	}
	for _, want := range []string{"Resource Inventory", "aws_security_group.web", "```mermaid", "graph TD"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestDocumentTerraformOutWritesFile(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	target := filepath.Join(t.TempDir(), "arch.md")
	if _, err := runCmd(t, "", "document", "terraform", reviewFixtureDir(t), "--out", target); err != nil {
		t.Fatalf("document: %v", err)
	}
	data, err := os.ReadFile(target) // #nosec G304 -- test-controlled temp path
	if err != nil || !strings.Contains(string(data), "Resource Inventory") {
		t.Fatalf("file not written: %v", err)
	}
}

func TestDocumentRunbookRequiresProvider(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	_, err := runCmd(t, "", "document", "terraform", reviewFixtureDir(t), "--type", "runbook")
	if err == nil || !strings.Contains(err.Error(), "atlas configure") {
		t.Fatalf("runbook without provider must point to configure: %v", err)
	}
}
