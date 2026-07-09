// Package review orchestrates infrastructure reviews: static analysis
// always runs; AI analysis is added when an adapter is available and
// degrades to warnings when it is not (offline-friendly by design).
package review

import (
	"context"

	"github.com/Haykhay/atlas/internal/aireview"
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

// aiAnalyze runs AI analysis and appends validated findings; every
// failure mode becomes a warning, never an error.
func aiAnalyze(
	ctx context.Context,
	ai provider.Adapter,
	read func() (map[string]string, error),
	system string,
	prompt func(map[string]string) string,
	findings *[]finding.Finding,
) []string {
	if ai == nil {
		return nil
	}
	files, err := read()
	if err != nil {
		return []string{"AI analysis skipped: " + err.Error()}
	}
	if len(files) == 0 {
		return nil
	}
	resp, err := ai.Complete(ctx, provider.Request{System: system, Prompt: prompt(files)})
	if err != nil {
		return []string{"AI analysis unavailable: " + err.Error()}
	}
	aiFindings, err := aireview.ParseFindings(resp.Text)
	if err != nil {
		return []string{"AI response could not be parsed: " + err.Error()}
	}
	*findings = append(*findings, aiFindings...)
	return nil
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
	warnings := aiAnalyze(ctx, t.AI,
		func() (map[string]string, error) { return terraform.ReadTFFiles(dir) },
		terraform.SystemPrompt, terraform.FilesPrompt, &findings)

	return Report{
		PillarScores: waf.Score(findings),
		Findings:     findings,
		Warnings:     warnings,
	}, nil
}
