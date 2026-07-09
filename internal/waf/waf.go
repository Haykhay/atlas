// Package waf models the AWS Well-Architected Framework pillars and
// scores findings against them.
package waf

import "github.com/Haykhay/atlas/internal/finding"

// The six Well-Architected pillars.
const (
	PillarOperationalExcellence = "Operational Excellence"
	PillarSecurity              = "Security"
	PillarReliability           = "Reliability"
	PillarPerformanceEfficiency = "Performance Efficiency"
	PillarCostOptimization      = "Cost Optimization"
	PillarSustainability        = "Sustainability"
)

// Pillars returns all six pillars in canonical order.
func Pillars() []string {
	return []string{
		PillarOperationalExcellence,
		PillarSecurity,
		PillarReliability,
		PillarPerformanceEfficiency,
		PillarCostOptimization,
		PillarSustainability,
	}
}

// PillarScore is the per-pillar result of a review.
type PillarScore struct {
	Pillar   string `json:"pillar"`
	Score    int    `json:"score"` // 0–100
	Findings int    `json:"findings"`
}

var deductions = map[string]int{
	finding.SeverityCritical: 25,
	finding.SeverityHigh:     15,
	finding.SeverityMedium:   8,
	finding.SeverityLow:      3,
	finding.SeverityInfo:     1,
}

// Score starts each pillar at 100 and deducts per finding by severity,
// flooring at 0. All six pillars are always returned.
func Score(findings []finding.Finding) []PillarScore {
	var out []PillarScore
	for _, pillar := range Pillars() {
		score, count := 100, 0
		for _, f := range findings {
			if f.Pillar != pillar {
				continue
			}
			count++
			score -= deductions[f.Severity]
		}
		if score < 0 {
			score = 0
		}
		out = append(out, PillarScore{Pillar: pillar, Score: score, Findings: count})
	}
	return out
}
