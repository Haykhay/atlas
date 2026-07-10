// Package cli defines the atlas command tree.
package cli

import (
	"github.com/spf13/cobra"

	// Register built-in providers.
	_ "github.com/Haykhay/atlas/internal/provider/anthropic"
	_ "github.com/Haykhay/atlas/internal/provider/ollama"
	_ "github.com/Haykhay/atlas/internal/provider/openai"
)

// Version is injected at build time via
// -ldflags "-X github.com/Haykhay/atlas/internal/cli.Version=v1.2.3".
var Version = "dev"

// NewRootCmd builds the atlas command tree. Tests construct their own
// instance; Execute() is the production entrypoint.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "atlas",
		Short:         "AI-native infrastructure engineering CLI",
		Long:          "Atlas reviews, explains, and documents cloud infrastructure using the AI provider of your choice.",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.AddCommand(newVersionCmd(), newProvidersCmd(), newConfigureCmd(), newReviewCmd(), newExplainCmd(), newDocumentCmd(), newFixCmd())
	return root
}

// Execute runs the atlas CLI with os.Args; it is the production entrypoint.
func Execute() error {
	return NewRootCmd().Execute()
}
