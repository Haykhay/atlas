package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Haykhay/atlas/internal/config"
	"github.com/Haykhay/atlas/internal/credentials"
)

func runCmd(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestProvidersListShowsRegisteredProviders(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	out, err := runCmd(t, "", "providers", "list")
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, name := range []string{"anthropic", "ollama", "openai"} {
		if !strings.Contains(out, name) {
			t.Errorf("list missing %q: %s", name, out)
		}
	}
}

func TestProvidersDefaultPersists(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	if _, err := runCmd(t, "", "providers", "default", "ollama"); err != nil {
		t.Fatalf("execute: %v", err)
	}
	cfg, err := config.Load()
	if err != nil || cfg.DefaultProvider != "ollama" {
		t.Fatalf("default not persisted: %+v, %v", cfg, err)
	}

	if _, err := runCmd(t, "", "providers", "default", "nope"); err == nil {
		t.Fatal("expected error for unknown provider")
	}
}

func TestProvidersLoginStoresCredentialAndLogoutRemoves(t *testing.T) {
	t.Setenv("ATLAS_CONFIG_DIR", t.TempDir())
	credentials.MockForTesting()

	if _, err := runCmd(t, "sk-test\n", "providers", "login", "anthropic"); err != nil {
		t.Fatalf("login: %v", err)
	}
	key, err := credentials.Get("anthropic")
	if err != nil || key != "sk-test" {
		t.Fatalf("credential not stored: %q, %v", key, err)
	}

	if _, err := runCmd(t, "", "providers", "logout", "anthropic"); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := credentials.Get("anthropic"); !credentials.IsNotFound(err) {
		t.Fatalf("credential not removed: %v", err)
	}
}
