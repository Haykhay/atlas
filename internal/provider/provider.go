// Package provider defines the AI provider abstraction. Every provider
// implements Adapter; the registry maps provider names to factories so
// the CLI stays provider-agnostic.
package provider

import (
	"context"
	"fmt"
	"sort"
)

// Request is a single completion request routed to any provider.
type Request struct {
	System    string
	Prompt    string
	MaxTokens int
}

// Response is the provider-neutral completion result.
type Response struct {
	Text  string
	Model string
}

// Adapter is implemented by every AI provider.
type Adapter interface {
	Name() string
	Complete(ctx context.Context, req Request) (Response, error)
	Status(ctx context.Context) error
}

// Settings carries everything an adapter needs at construction time.
// APIKey comes from the OS keychain; BaseURL and Model from config.
type Settings struct {
	APIKey  string
	BaseURL string
	Model   string
}

// Factory constructs an adapter from settings.
type Factory func(s Settings) Adapter

var registry = map[string]Factory{}

// Register makes a provider available by name; adapter packages call
// this from init().
func Register(name string, f Factory) {
	registry[name] = f
}

// Names returns registered provider names, sorted.
func Names() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// New constructs the named adapter.
func New(name string, s Settings) (Adapter, error) {
	f, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider %q (known: %v)", name, Names())
	}
	return f(s), nil
}
