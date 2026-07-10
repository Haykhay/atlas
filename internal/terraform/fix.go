package terraform

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Haykhay/atlas/internal/aireview"
	"github.com/Haykhay/atlas/internal/finding"
)

// FixSystemPrompt instructs the model to emit a unified diff only.
// Atlas never applies changes — the diff is the deliverable.
const FixSystemPrompt = `You are Atlas, an infrastructure engineering tool. Produce fixes for the listed findings as ONE unified diff (git apply compatible) against the provided files.

Rules:
- Respond with ONLY the unified diff inside a single ` + "```diff" + ` code fence.
- Use "--- a/<file>" / "+++ b/<file>" headers with correct hunk offsets.
- Change only what the findings require; preserve formatting elsewhere.
- If a finding cannot be fixed from the given files, skip it silently.`

// FixPrompt renders the findings and files for the model.
func FixPrompt(files map[string]string, findings []finding.Finding) string {
	var b strings.Builder
	b.WriteString("Findings to fix:\n")
	for _, f := range findings {
		fmt.Fprintf(&b, "- %s: %s — %s\n", f.ID, f.Title, f.Remediation)
	}
	b.WriteString("\n")
	b.WriteString(aireview.FilesPrompt("Files to patch:", files))
	return b.String()
}

// ExtractDiff pulls the unified diff out of a model response: a
// ```diff fence if present, otherwise the first "--- " block.
func ExtractDiff(text string) (string, error) {
	if i := strings.Index(text, "```diff"); i >= 0 {
		rest := text[i+len("```diff"):]
		if j := strings.Index(rest, "```"); j >= 0 {
			return strings.TrimSpace(rest[:j]) + "\n", nil
		}
	}
	if i := strings.Index(text, "--- "); i >= 0 {
		diff := text[i:]
		if j := strings.Index(diff, "```"); j >= 0 {
			diff = diff[:j]
		}
		return strings.TrimSpace(diff) + "\n", nil
	}
	return "", errors.New("no unified diff in AI response")
}
