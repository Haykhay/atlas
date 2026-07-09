package cli

import (
	"github.com/spf13/cobra"
)

// Version is injected at build time via
// -ldflags "-X github.com/silverstreaminnovations/atlas/internal/cli.Version=v1.2.3".
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	root.AddCommand(newVersionCmd())
	return root
}

func Execute() error {
	return NewRootCmd().Execute()
}
