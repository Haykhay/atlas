package redact

import (
	"strings"
	"testing"
)

func TestRedactDetectsEachKind(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		secret string // must not survive
		kind   string
	}{
		{"aws access key", `provider "aws" { access_key = "AKIAIOSFODNN7EXAMPLE" }`, "AKIAIOSFODNN7EXAMPLE", "api_key"},
		{"anthropic key", `key: sk-ant-api03-abcdefghij1234567890xyz`, "sk-ant-api03-abcdefghij1234567890xyz", "api_key"},
		{"api key assignment", `api_key = "9f8e7d6c5b4a3210ffee"`, "9f8e7d6c5b4a3210ffee", "api_key"},
		{"github token", `remote: https://ghp_AbCdEfGhIjKlMnOpQrStUvWxYz123456@github.com`, "ghp_AbCdEfGhIjKlMnOpQrStUvWxYz123456", "token"},
		{"slack token", `SLACK=xoxb-1234567890-abcdefghij`, "xoxb-1234567890-abcdefghij", "token"},
		{"bearer header", `Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.payload`, "eyJhbGciOiJIUzI1NiJ9.payload", "token"},
		{"token assignment", `access_token: 0123456789abcdef`, "0123456789abcdef", "token"},
		{"password", `password = "hunter22secret"`, "hunter22secret", "password"},
		{"secret", `aws_secret_access_key = wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY`, "wJalrXUtnFEMIK7MDENGbPxRfiCYEXAMPLEKEY", "secret"},
		{"arn account id", `arn:aws:iam::123456789012:role/deploy`, "123456789012", "account_id"},
		{"account id assignment", `account_id = "210987654321"`, "210987654321", "account_id"},
		{"internal domain url", `curl https://vault.corp/v1/kv`, "https://vault.corp/v1/kv", "internal_url"},
		{"rfc1918 url", `endpoint http://10.0.12.7:8500/ui`, "http://10.0.12.7:8500/ui", "internal_url"},
		{"private key block", "-----BEGIN RSA PRIVATE KEY-----\nMIIEow\n-----END RSA PRIVATE KEY-----", "MIIEow", "private_key"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			clean, report := Redact(tc.input)
			if strings.Contains(clean, tc.secret) {
				t.Fatalf("secret survived redaction: %s", clean)
			}
			if len(report) == 0 {
				t.Fatal("expected a redaction report entry")
			}
			found := false
			for _, r := range report {
				if r.Kind == tc.kind {
					found = true
				}
			}
			if !found {
				t.Fatalf("expected kind %q in report %+v", tc.kind, report)
			}
			if !strings.Contains(clean, "«REDACTED:"+tc.kind) {
				t.Fatalf("expected placeholder for %q in output: %s", tc.kind, clean)
			}
		})
	}
}

func TestRedactLeavesCleanTextUntouched(t *testing.T) {
	in := `resource "aws_s3_bucket" "logs" { bucket = "example-logs" }`
	clean, report := Redact(in)
	if clean != in {
		t.Fatalf("clean text was modified: %s", clean)
	}
	if len(report) != 0 {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestRedactNumbersPlaceholdersPerKind(t *testing.T) {
	in := "password = firstpass99\npassword = secondpass99"
	clean, report := Redact(in)
	if !strings.Contains(clean, "«REDACTED:password:1»") || !strings.Contains(clean, "«REDACTED:password:2»") {
		t.Fatalf("expected numbered placeholders: %s", clean)
	}
	if len(report) != 2 {
		t.Fatalf("expected 2 report entries, got %+v", report)
	}
}
