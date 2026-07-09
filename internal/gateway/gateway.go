// Package gateway wraps provider adapters with the redaction engine.
// Review engines must always talk to a Gateway, never a raw adapter —
// that is what guarantees no sensitive content leaves the machine.
package gateway

import (
	"context"

	"github.com/Haykhay/atlas/internal/provider"
	"github.com/Haykhay/atlas/internal/redact"
)

// Gateway is a provider.Adapter that redacts request content before
// delegating to the wrapped adapter.
type Gateway struct {
	inner provider.Adapter

	// OnRedact, when set, is called with the redaction report whenever
	// a request had sensitive content removed (e.g. to tell the user
	// "3 secrets were redacted before sending").
	OnRedact func([]redact.Redaction)
}

// New wraps an adapter with mandatory outbound redaction.
func New(inner provider.Adapter) *Gateway {
	return &Gateway{inner: inner}
}

// Name implements provider.Adapter.
func (g *Gateway) Name() string { return g.inner.Name() }

// Complete implements provider.Adapter, redacting System and Prompt
// before anything is sent.
func (g *Gateway) Complete(ctx context.Context, req provider.Request) (provider.Response, error) {
	var all []redact.Redaction

	clean, rep := redact.Redact(req.System)
	req.System = clean
	all = append(all, rep...)

	clean, rep = redact.Redact(req.Prompt)
	req.Prompt = clean
	all = append(all, rep...)

	if g.OnRedact != nil && len(all) > 0 {
		g.OnRedact(all)
	}
	return g.inner.Complete(ctx, req)
}

// Status implements provider.Adapter.
func (g *Gateway) Status(ctx context.Context) error {
	return g.inner.Status(ctx)
}
