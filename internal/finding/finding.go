// Package finding defines the Atlas trust-layer contract: every
// recommendation Atlas emits is a Finding carrying evidence, confidence,
// severity, business impact, remediation, and its origin. A finding
// without evidence is invalid by design.
package finding

import (
	"errors"
	"fmt"
)

// Origin records how a finding was produced.
type Origin string

// Valid origins.
const (
	OriginStatic Origin = "static"
	OriginAI     Origin = "ai"
	OriginBoth   Origin = "both"
)

// Severity levels, most severe first.
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)

// Finding is the trust-layer contract locked in the V1 roadmap.
type Finding struct {
	ID                string   `json:"id"`
	Title             string   `json:"title"`
	Severity          string   `json:"severity"`   // critical|high|medium|low|info
	Confidence        float64  `json:"confidence"` // 0.0–1.0
	Evidence          []string `json:"evidence"`   // file:line excerpts; never empty
	AffectedResources []string `json:"affected_resources"`
	BusinessImpact    string   `json:"business_impact"`
	Remediation       string   `json:"remediation"`
	Patch             string   `json:"patch,omitempty"` // unified diff, never auto-applied
	Origin            Origin   `json:"origin"`
	Pillar            string   `json:"pillar,omitempty"` // WAF pillar, terraform reviews only
	References        []string `json:"references,omitempty"`
}

var validSeverities = map[string]bool{
	SeverityCritical: true,
	SeverityHigh:     true,
	SeverityMedium:   true,
	SeverityLow:      true,
	SeverityInfo:     true,
}

var validOrigins = map[Origin]bool{
	OriginStatic: true,
	OriginAI:     true,
	OriginBoth:   true,
}

// Validate enforces the trust-layer invariants; a Finding that fails
// here must never be shown to a user or emitted by a command.
func (f Finding) Validate() error {
	if f.ID == "" {
		return errors.New("finding: ID is required")
	}
	if f.Title == "" {
		return errors.New("finding: Title is required")
	}
	if !validSeverities[f.Severity] {
		return fmt.Errorf("finding %s: invalid severity %q", f.ID, f.Severity)
	}
	if f.Confidence < 0 || f.Confidence > 1 {
		return fmt.Errorf("finding %s: confidence %v outside [0,1]", f.ID, f.Confidence)
	}
	if len(f.Evidence) == 0 {
		return fmt.Errorf("finding %s: at least one piece of evidence is required", f.ID)
	}
	if !validOrigins[f.Origin] {
		return fmt.Errorf("finding %s: invalid origin %q", f.ID, f.Origin)
	}
	return nil
}
