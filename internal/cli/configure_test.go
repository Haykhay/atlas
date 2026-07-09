package cli

import (
	"strings"
	"testing"

	"github.com/silverstreaminnovations/atlas/internal/config"
	"github.com/silverstreaminnovations/atlas/internal/credentials"
)

func TestConfigureSelectsProviderAndSetsDefault(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	// provider.Names() is sorted: anthropic(1), ollama(2), openai(3).
	// Choose ollama, accept default base URL.
	out, err := runCmd(t, "2\n\n", "configure")
	if err != nil {
		t.Fatalf("configure: %v", err)
	}
	if !strings.Contains(out, "ollama") {
		t.Fatalf("expected provider menu, got: %s", out)
	}
	cfg, err := config.Load()
	if err != nil || cfg.DefaultProvider != "ollama" {
		t.Fatalf("default not set: %+v, %v", cfg, err)
	}
}

func TestConfigureRejectsInvalidChoice(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	if _, err := runCmd(t, "99\n", "configure"); err == nil {
		t.Fatal("expected error for invalid selection")
	}
}
