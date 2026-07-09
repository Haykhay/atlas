package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/silverstreaminnovations/atlas/internal/config"
	"github.com/silverstreaminnovations/atlas/internal/credentials"
	"github.com/silverstreaminnovations/atlas/internal/provider"
)

func newProvidersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "Manage AI providers",
	}
	cmd.AddCommand(
		newProvidersListCmd(),
		newProvidersDefaultCmd(),
		newProvidersLoginCmd(),
		newProvidersLogoutCmd(),
		newProvidersStatusCmd(),
	)
	return cmd
}

func newProvidersListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available providers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			for _, name := range provider.Names() {
				marker := " "
				if name == cfg.DefaultProvider {
					marker = "*"
				}
				state := "not configured"
				if name == "ollama" {
					state = "local (no credentials needed)"
				} else if _, err := credentials.Get(name); err == nil {
					state = "configured"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%s %-12s %s\n", marker, name, state)
			}
			return nil
		},
	}
}

func newProvidersDefaultCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "default <provider>",
		Short: "Set the default provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if !known(name) {
				return fmt.Errorf("unknown provider %q (known: %v)", name, provider.Names())
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			cfg.DefaultProvider = name
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Default provider set to %s.\n", name)
			return nil
		},
	}
}

func newProvidersLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login <provider>",
		Short: "Store credentials for a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if err := loginProvider(cmd, cfg, args[0]); err != nil {
				return err
			}
			return config.Save(cfg)
		},
	}
}

func newProvidersLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout <provider>",
		Short: "Remove stored credentials for a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := credentials.Delete(args[0])
			if err != nil && !credentials.IsNotFound(err) {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Logged out of %s.\n", args[0])
			return nil
		},
	}
}

func newProvidersStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [provider]",
		Short: "Check provider connectivity",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			names := provider.Names()
			if len(args) == 1 {
				if !known(args[0]) {
					return fmt.Errorf("unknown provider %q (known: %v)", args[0], provider.Names())
				}
				names = args[:1]
			}
			for _, name := range names {
				fmt.Fprintf(cmd.OutOrStdout(), "%-12s %s\n", name, statusLine(cmd.Context(), cfg, name))
			}
			return nil
		},
	}
}

func statusLine(ctx context.Context, cfg *config.Config, name string) string {
	key, err := credentials.Get(name)
	if err != nil && !credentials.IsNotFound(err) {
		return "error: " + err.Error()
	}
	if key == "" && name != "ollama" {
		return "not configured (run: atlas providers login " + name + ")"
	}
	pc := cfg.Providers[name]
	a, err := provider.New(name, provider.Settings{APIKey: key, BaseURL: pc.BaseURL, Model: pc.Model})
	if err != nil {
		return "error: " + err.Error()
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := a.Status(ctx); err != nil {
		return "unreachable: " + err.Error()
	}
	return "ok"
}

// loginProvider prompts for and stores the secrets/settings a provider
// needs. Shared by `providers login` and `configure`.
func loginProvider(cmd *cobra.Command, cfg *config.Config, name string) error {
	if !known(name) {
		return fmt.Errorf("unknown provider %q (known: %v)", name, provider.Names())
	}
	if name == "ollama" {
		fmt.Fprint(cmd.OutOrStdout(), "Ollama base URL [http://localhost:11434]: ")
		base, err := readLine(cmd)
		if err != nil {
			return err
		}
		pc := cfg.Providers[name]
		pc.BaseURL = base
		cfg.Providers[name] = pc
		fmt.Fprintln(cmd.OutOrStdout(), "Ollama configured.")
		return nil
	}
	fmt.Fprintf(cmd.OutOrStdout(), "API key for %s: ", name)
	key, err := readSecret(cmd)
	if err != nil {
		return err
	}
	if key == "" {
		return errors.New("no API key entered")
	}
	if err := credentials.Set(name, key); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Credentials for %s stored in the OS keychain.\n", name)
	return nil
}

func known(name string) bool {
	for _, n := range provider.Names() {
		if n == name {
			return true
		}
	}
	return false
}

func readLine(cmd *cobra.Command) (string, error) {
	line, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

// readSecret reads without echo when stdin is a terminal, so API keys
// are never displayed; in tests (non-tty) it falls back to readLine.
func readSecret(cmd *cobra.Command) (string, error) {
	if f, ok := cmd.InOrStdin().(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(cmd.OutOrStdout())
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	return readLine(cmd)
}
