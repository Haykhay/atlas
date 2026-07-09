package review

import (
	"context"

	"github.com/Haykhay/atlas/internal/kubernetes"
	"github.com/Haykhay/atlas/internal/provider"
	"github.com/Haykhay/atlas/internal/waf"
)

// Kubernetes reviews manifest directories. AI must already be wrapped
// in the redaction gateway by the caller; nil means static-only.
type Kubernetes struct {
	AI provider.Adapter
}

// Run parses dir, applies static rules, optionally merges AI findings,
// and scores the six Well-Architected pillars.
func (k *Kubernetes) Run(ctx context.Context, dir string) (Report, error) {
	objects, err := kubernetes.ParseDir(dir)
	if err != nil {
		return Report{}, err
	}
	findings := kubernetes.RunRules(objects)
	warnings := aiAnalyze(ctx, k.AI,
		func() (map[string]string, error) { return kubernetes.ReadManifests(dir) },
		kubernetes.SystemPrompt, kubernetes.FilesPrompt, &findings)

	return Report{
		PillarScores: waf.Score(findings),
		Findings:     findings,
		Warnings:     warnings,
	}, nil
}
