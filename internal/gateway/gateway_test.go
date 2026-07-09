package gateway

import (
	"context"
	"strings"
	"testing"

	"github.com/Haykhay/atlas/internal/provider"
	"github.com/Haykhay/atlas/internal/redact"
)

type capturingAdapter struct {
	got         provider.Request
	statusCalls int
}

func (c *capturingAdapter) Name() string { return "capture" }
func (c *capturingAdapter) Complete(_ context.Context, req provider.Request) (provider.Response, error) {
	c.got = req
	return provider.Response{Text: "ok", Model: "m"}, nil
}
func (c *capturingAdapter) Status(_ context.Context) error {
	c.statusCalls++
	return nil
}

var _ provider.Adapter = (*Gateway)(nil)

func TestCompleteRedactsBeforeDelegating(t *testing.T) {
	inner := &capturingAdapter{}
	g := New(inner)

	var reported []redact.Redaction
	g.OnRedact = func(r []redact.Redaction) { reported = r }

	_, err := g.Complete(context.Background(), provider.Request{
		System: "review this",
		Prompt: `variable "db" { default = "password = hunter22secret" }`,
	})
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if strings.Contains(inner.got.Prompt, "hunter22secret") {
		t.Fatalf("secret leaked to adapter: %s", inner.got.Prompt)
	}
	if !strings.Contains(inner.got.Prompt, "«REDACTED:password:1»") {
		t.Fatalf("expected placeholder in forwarded prompt: %s", inner.got.Prompt)
	}
	if len(reported) != 1 || reported[0].Kind != "password" {
		t.Fatalf("unexpected redaction report: %+v", reported)
	}
}

func TestCleanRequestPassesThroughWithoutCallback(t *testing.T) {
	inner := &capturingAdapter{}
	g := New(inner)

	called := false
	g.OnRedact = func([]redact.Redaction) { called = true }

	prompt := `resource "aws_s3_bucket" "logs" {}`
	if _, err := g.Complete(context.Background(), provider.Request{Prompt: prompt}); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if inner.got.Prompt != prompt {
		t.Fatalf("clean prompt modified: %s", inner.got.Prompt)
	}
	if called {
		t.Fatal("OnRedact must not fire for clean requests")
	}
}

func TestNameAndStatusDelegate(t *testing.T) {
	inner := &capturingAdapter{}
	g := New(inner)
	if g.Name() != "capture" {
		t.Fatalf("name: %s", g.Name())
	}
	if err := g.Status(context.Background()); err != nil || inner.statusCalls != 1 {
		t.Fatalf("status not delegated: %v, calls=%d", err, inner.statusCalls)
	}
}
