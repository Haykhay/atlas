// Package redact strips secrets, credentials, and sensitive identifiers
// from text before it leaves the developer's machine. Every outbound
// provider request must pass through Redact (see internal/gateway).
package redact

import (
	"fmt"
	"regexp"
)

// Redaction describes one replaced span. The original value is never
// stored — the mapping stays inside the process.
type Redaction struct {
	Kind        string // secret|api_key|token|password|private_key|account_id|internal_url|metadata
	Placeholder string // e.g. "«REDACTED:account_id:1»"
}

// rule pairs a detector with the redaction kind. When group is 1, only
// the captured value is replaced (keeps surrounding syntax readable);
// when 0, the whole match is replaced.
type rule struct {
	kind  string
	re    *regexp.Regexp
	group int
}

// Value character classes exclude « and » so placeholders can never
// rematch a rule. private_key runs first so generic assignment rules
// don't shred key blocks.
var rules = []rule{
	{kind: "private_key", re: regexp.MustCompile(`(?s)-----BEGIN [A-Z ]*PRIVATE KEY-----.*?-----END [A-Z ]*PRIVATE KEY-----`)},
	{kind: "api_key", re: regexp.MustCompile(`\b(?:AKIA|ASIA)[0-9A-Z]{16}\b`)},
	{kind: "api_key", re: regexp.MustCompile(`\bsk-(?:ant-)?[A-Za-z0-9_\-]{20,}`)},
	{kind: "token", re: regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{20,}\b`)},
	{kind: "token", re: regexp.MustCompile(`\bxox[baprs]-[A-Za-z0-9\-]{10,}\b`)},
	{kind: "token", re: regexp.MustCompile(`(?i)\bbearer\s+([A-Za-z0-9._\-]{16,})`), group: 1},
	{kind: "password", re: regexp.MustCompile(`(?i)\b(?:password|passwd|pwd)\b["']?\s*[:=]\s*["']?([^\s"'«»]{4,})`), group: 1},
	{kind: "secret", re: regexp.MustCompile(`(?i)\b(?:aws_secret_access_key|client_secret|secret_key|secret)\b["']?\s*[:=]\s*["']?([^\s"'«»]{4,})`), group: 1},
	{kind: "api_key", re: regexp.MustCompile(`(?i)\bapi[_-]?key\b["']?\s*[:=]\s*["']?([^\s"'«»]{4,})`), group: 1},
	{kind: "token", re: regexp.MustCompile(`(?i)\b(?:auth[_-]?token|access[_-]?token|token)\b["']?\s*[:=]\s*["']?([^\s"'«»]{4,})`), group: 1},
	{kind: "account_id", re: regexp.MustCompile(`\barn:aws[a-z0-9\-]*:[^:\s]*:[^:\s]*:(\d{12}):`), group: 1},
	{kind: "account_id", re: regexp.MustCompile(`(?i)\baccount[_-]?id\b["']?\s*[:=]\s*["']?(\d{12})\b`), group: 1},
	{kind: "internal_url", re: regexp.MustCompile(`https?://[A-Za-z0-9.\-]+\.(?:internal|local|corp|intranet)(?::\d+)?[^\s"']*`)},
	{kind: "internal_url", re: regexp.MustCompile(`https?://(?:10\.\d{1,3}\.\d{1,3}\.\d{1,3}|192\.168\.\d{1,3}\.\d{1,3}|172\.(?:1[6-9]|2\d|3[01])\.\d{1,3}\.\d{1,3})(?::\d+)?[^\s"']*`)},
}

// maxPerRule bounds the replace loop as defense-in-depth against any
// pathological rematch; real inputs stay far below it.
const maxPerRule = 500

// Redact replaces sensitive spans with «REDACTED:kind:N» placeholders
// and reports what was removed (kinds only, never values).
func Redact(input string) (string, []Redaction) {
	out := input
	counts := map[string]int{}
	var report []Redaction

	for _, r := range rules {
		for i := 0; i < maxPerRule; i++ {
			loc := r.re.FindStringSubmatchIndex(out)
			if loc == nil {
				break
			}
			start, end := loc[0], loc[1]
			if r.group > 0 {
				start, end = loc[2*r.group], loc[2*r.group+1]
			}
			counts[r.kind]++
			ph := fmt.Sprintf("«REDACTED:%s:%d»", r.kind, counts[r.kind])
			out = out[:start] + ph + out[end:]
			report = append(report, Redaction{Kind: r.kind, Placeholder: ph})
		}
	}
	return out, report
}
