package cli

import (
	"strings"
	"testing"

	"github.com/Haykhay/atlas/internal/credentials"
)

func TestExplainRequiresConfiguredProvider(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	_, err := runCmd(t, "", "explain", "terraform", t.TempDir())
	if err == nil {
		t.Fatal("explain without a provider must fail")
	}
	if !strings.Contains(err.Error(), "atlas configure") {
		t.Fatalf("error should point to atlas configure: %v", err)
	}
}
