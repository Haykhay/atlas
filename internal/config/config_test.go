package config

import (
	"testing"
)

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.DefaultProvider != "" {
		t.Fatalf("expected empty default provider, got %q", cfg.DefaultProvider)
	}
	if cfg.Providers == nil {
		t.Fatal("Providers map must be initialized")
	}
}

func TestSaveThenLoadRoundTrip(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())

	in := &Config{
		DefaultProvider: "ollama",
		Providers: map[string]ProviderConfig{
			"ollama": {Model: "llama3"},
		},
	}
	if err := Save(in); err != nil {
		t.Fatalf("save: %v", err)
	}

	out, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if out.DefaultProvider != "ollama" {
		t.Fatalf("default provider: got %q, want %q", out.DefaultProvider, "ollama")
	}
	if out.Providers["ollama"].Model != "llama3" {
		t.Fatalf("model: got %q, want %q", out.Providers["ollama"].Model, "llama3")
	}
}
