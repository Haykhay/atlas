// Package aireview holds the AI plumbing shared by every review
// domain: rendering configuration files into a prompt and parsing the
// model's JSON findings back into validated trust-layer findings.
package aireview

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Haykhay/atlas/internal/finding"
)

// FindingContract is the response contract embedded in every review
// system prompt. The trust layer re-validates locally — the model is
// never trusted to validate itself.
const FindingContract = `Respond with ONLY a JSON array of findings. Each finding must have exactly these fields:
- "id": string, format "ATLAS-AI-NNN"
- "title": short defect statement
- "severity": one of "critical","high","medium","low","info"
- "confidence": number 0.0-1.0
- "evidence": array of "file:line: <exact quoted source>" strings (required, never empty)
- "affected_resources": array of resource addresses
- "business_impact": one sentence
- "remediation": one sentence
- "pillar": one of "Operational Excellence","Security","Reliability","Performance Efficiency","Cost Optimization","Sustainability"
- "references": array of official documentation URLs

Report only real, defensible issues visible in the provided source. If there are none, respond with [].`

// FilesPrompt renders intro plus the files, sorted by name for
// determinism, each fenced for the model.
func FilesPrompt(intro string, files map[string]string) string {
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString(intro)
	b.WriteString("\n")
	for _, n := range names {
		fmt.Fprintf(&b, "\n## %s\n```\n%s\n```\n", n, files[n])
	}
	return b.String()
}

// ParseFindings extracts the JSON findings array from a model response,
// forces origin=ai, normalizes confidence, and drops anything that
// fails trust-layer validation.
func ParseFindings(text string) ([]finding.Finding, error) {
	start := strings.Index(text, "[")
	end := strings.LastIndex(text, "]")
	if start < 0 || end <= start {
		return nil, errors.New("no JSON array in AI response")
	}
	var raw []finding.Finding
	if err := json.Unmarshal([]byte(text[start:end+1]), &raw); err != nil {
		return nil, fmt.Errorf("decode AI findings: %w", err)
	}

	valid := make([]finding.Finding, 0, len(raw))
	for _, f := range raw {
		f.Origin = finding.OriginAI
		if f.Confidence > 1 {
			f.Confidence /= 100
		}
		if f.Confidence > 1 {
			f.Confidence = 1
		}
		if f.Confidence < 0 {
			f.Confidence = 0
		}
		if err := f.Validate(); err != nil {
			continue
		}
		valid = append(valid, f)
	}
	return valid, nil
}
