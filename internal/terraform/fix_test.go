package terraform

import (
	"strings"
	"testing"

	"github.com/Haykhay/atlas/internal/finding"
)

func TestExtractDiffFromFence(t *testing.T) {
	text := "Here is the fix:\n```diff\n--- a/main.tf\n+++ b/main.tf\n@@ -1 +1 @@\n-storage_encrypted = false\n+storage_encrypted = true\n```\nApply with git apply."
	diff, err := ExtractDiff(text)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if !strings.HasPrefix(diff, "--- a/main.tf") || strings.Contains(diff, "```") {
		t.Fatalf("unexpected diff: %q", diff)
	}
}

func TestExtractDiffBare(t *testing.T) {
	text := "--- a/main.tf\n+++ b/main.tf\n@@ -1 +1 @@\n-x\n+y\n"
	diff, err := ExtractDiff(text)
	if err != nil || !strings.Contains(diff, "+++ b/main.tf") {
		t.Fatalf("bare diff not extracted: %q, %v", diff, err)
	}
}

func TestExtractDiffMissing(t *testing.T) {
	if _, err := ExtractDiff("I have no changes to suggest."); err == nil {
		t.Fatal("expected error when no diff present")
	}
}

func TestFixPromptIncludesFindingsAndFiles(t *testing.T) {
	p := FixPrompt(
		map[string]string{"main.tf": `resource "aws_db_instance" "app" {}`},
		[]finding.Finding{{ID: "ATLAS-TF-003", Title: "RDS instance without storage encryption", Remediation: "Set storage_encrypted = true"}},
	)
	for _, want := range []string{"ATLAS-TF-003", "storage_encrypted = true", "main.tf", "aws_db_instance"} {
		if !strings.Contains(p, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}
