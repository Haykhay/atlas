package waf

import (
	"testing"

	"github.com/Haykhay/atlas/internal/finding"
)

func TestScoreWithNoFindingsIsPerfect(t *testing.T) {
	scores := Score(nil)
	if len(scores) != 6 {
		t.Fatalf("expected 6 pillars, got %d", len(scores))
	}
	for _, s := range scores {
		if s.Score != 100 || s.Findings != 0 {
			t.Errorf("pillar %s: expected 100/0, got %d/%d", s.Pillar, s.Score, s.Findings)
		}
	}
}

func TestScoreDeductsBySeverity(t *testing.T) {
	scores := Score([]finding.Finding{
		{Severity: finding.SeverityCritical, Pillar: PillarSecurity},
		{Severity: finding.SeverityLow, Pillar: PillarSecurity},
	})
	for _, s := range scores {
		switch s.Pillar {
		case PillarSecurity:
			if s.Score != 72 || s.Findings != 2 { // 100 - 25 - 3
				t.Errorf("security: got %d/%d", s.Score, s.Findings)
			}
		default:
			if s.Score != 100 {
				t.Errorf("%s should be untouched, got %d", s.Pillar, s.Score)
			}
		}
	}
}

func TestScoreFloorsAtZero(t *testing.T) {
	var fs []finding.Finding
	for i := 0; i < 10; i++ {
		fs = append(fs, finding.Finding{Severity: finding.SeverityCritical, Pillar: PillarReliability})
	}
	for _, s := range Score(fs) {
		if s.Pillar == PillarReliability && s.Score != 0 {
			t.Fatalf("expected floor 0, got %d", s.Score)
		}
	}
}
