// Package review orchestrates infrastructure reviews: static analysis
// always runs; AI analysis is added when an adapter is available and
// degrades to warnings when it is not (offline-friendly by design).
package review

import (
	"context"

	"github.com/Haykhay/atlas/internal/finding"
	"github.com/Haykhay/atlas/internal/provider"
	"github.com/Haykhay/atlas/internal/terraform"
	"github.com/Haykhay/atlas/internal/waf"
)

// Report is the result of a review run.
type Report struct {
	PillarScores []waf.PillarScore `json:"pillar_scores"`
	Findings     []finding.Finding `json:"findings"`
	Warnings     []string          `json:"warnings,omitempty"`
}

// Terraform reviews Terraform directories. AI must already be wrapped
// in the redaction gateway by the caller; nil means static-only.
type Terraform struct {
	AI provider.Adapter
}

// Run parses dir, applies static rules, optionally merges AI findings,
// and scores the six Well-Architected pillars.
func (t *Terraform) Run(ctx context.Context, dir string) (Report, error) {
	resources, err := terraform.ParseDir(dir)
	if err != nil {
		return Report{}, err
	}
	findings := terraform.RunRules(resources)
	var warnings []string

	if t.AI != nil {
		files, err := terraform.ReadTFFiles(dir)
		switch {
		case err != nil:
			warnings = append(warnings, "AI analysis skipped: "+err.Error())
		case len(files) == 0:
			// nothing to analyze
		default:
			resp, err := t.AI.Complete(ctx, provider.Request{
				System: terraform.SystemPrompt,
				Prompt: terraform.FilesPrompt(files),
			})
			if err != nil {
				warnings = append(warnings, "AI analysis unavailable: "+err.Error())
				break
			}
			aiFindings, err := terraform.ParseAIFindings(resp.Text)
			if err != nil {
				warnings = append(warnings, "AI response could not be parsed: "+err.Error())
				break
			}
			findings = append(findings, aiFindings...)
		}
	}

	return Report{
		PillarScores: waf.Score(findings),
		Findings:     findings,
		Warnings:     warnings,
	}, nil
}
