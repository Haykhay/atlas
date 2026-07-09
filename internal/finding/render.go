package finding

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

var severityRank = map[string]int{
	SeverityCritical: 0,
	SeverityHigh:     1,
	SeverityMedium:   2,
	SeverityLow:      3,
	SeverityInfo:     4,
}

// RenderJSON writes findings as an indented JSON array (stable format
// for CI/CD and tooling).
func RenderJSON(w io.Writer, findings []Finding) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(findings)
}

// RenderHuman writes findings for terminal reading, most severe first.
func RenderHuman(w io.Writer, findings []Finding) error {
	sorted := make([]Finding, len(findings))
	copy(sorted, findings)
	sort.SliceStable(sorted, func(i, j int) bool {
		return severityRank[sorted[i].Severity] < severityRank[sorted[j].Severity]
	})

	for _, f := range sorted {
		if _, err := fmt.Fprintf(w, "[%s] %s (%s, confidence %.0f%%)\n",
			strings.ToUpper(f.Severity), f.Title, f.ID, f.Confidence*100); err != nil {
			return err
		}
		if len(f.AffectedResources) > 0 {
			fmt.Fprintf(w, "  Resources:   %s\n", strings.Join(f.AffectedResources, ", "))
		}
		for _, e := range f.Evidence {
			fmt.Fprintf(w, "  Evidence:    %s\n", e)
		}
		if f.BusinessImpact != "" {
			fmt.Fprintf(w, "  Impact:      %s\n", f.BusinessImpact)
		}
		if f.Remediation != "" {
			fmt.Fprintf(w, "  Remediation: %s\n", f.Remediation)
		}
		origin := string(f.Origin)
		if f.Pillar != "" {
			origin += " | pillar: " + f.Pillar
		}
		fmt.Fprintf(w, "  Origin:      %s\n", origin)
		for _, ref := range f.References {
			fmt.Fprintf(w, "  Reference:   %s\n", ref)
		}
		fmt.Fprintln(w)
	}
	return nil
}
