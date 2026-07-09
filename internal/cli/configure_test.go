package cli

import (
	"fmt"
	"strings"
	"testing"

	"github.com/Haykhay/atlas/internal/config"
	"github.com/Haykhay/atlas/internal/credentials"
	"github.com/Haykhay/atlas/internal/provider"
)

func TestConfigureSelectsProviderAndSetsDefault(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	// Choose ollama by its menu position, then accept the default base URL.
	choice := 0
	for i, n := range provider.Names() {
		if n == "ollama" {
			choice = i + 1
		}
	}
	out, err := runCmd(t, fmt.Sprintf("%d\n\n", choice), "configure")
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
