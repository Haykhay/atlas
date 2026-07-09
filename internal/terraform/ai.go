package terraform

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/Haykhay/atlas/internal/finding"
)

// SystemPrompt instructs the model to behave as a reviewer and reply
// with machine-parseable findings only. The trust-layer contract is
// enforced again locally by ParseAIFindings — the model is never
// trusted to validate itself.
const SystemPrompt = `You are Atlas, an infrastructure engineering review tool. Review the provided Terraform configuration against the six AWS Well-Architected pillars (Operational Excellence, Security, Reliability, Performance Efficiency, Cost Optimization, Sustainability).

Respond with ONLY a JSON array of findings. Each finding must have exactly these fields:
- "id": string, format "ATLAS-AI-NNN"
- "title": short defect statement
- "severity": one of "critical","high","medium","low","info"
- "confidence": number 0.0-1.0
- "evidence": array of "file:line: <exact quoted source>" strings (required, never empty)
- "affected_resources": array of terraform addresses like "aws_s3_bucket.logs"
- "business_impact": one sentence
- "remediation": one sentence
- "pillar": the Well-Architected pillar name
- "references": array of AWS documentation URLs

Report only real, defensible issues visible in the provided source. If there are none, respond with [].`

// FilesPrompt renders the configuration files for the model, sorted by
// name for determinism.
func FilesPrompt(files map[string]string) string {
	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)

	var b strings.Builder
	b.WriteString("Terraform configuration to review:\n")
	for _, n := range names {
		fmt.Fprintf(&b, "\n## %s\n```hcl\n%s\n```\n", n, files[n])
	}
	return b.String()
}

// ParseAIFindings extracts the JSON findings array from a model
// response, forces origin=ai, normalizes confidence, and drops anything
// that fails trust-layer validation.
func ParseAIFindings(text string) ([]finding.Finding, error) {
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
