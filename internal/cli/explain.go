package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Haykhay/atlas/internal/config"
	"github.com/Haykhay/atlas/internal/gateway"
	"github.com/Haykhay/atlas/internal/provider"
	"github.com/Haykhay/atlas/internal/redact"
	"github.com/Haykhay/atlas/internal/terraform"
)

const explainSystemPrompt = `You are Atlas, an infrastructure engineering assistant. Explain Terraform configurations accurately using correct cloud terminology. Structure the explanation by resource, describe how the pieces connect, and call out anything surprising. Do not invent resources that are not in the configuration.`

func newExplainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain",
		Short: "Explain infrastructure in plain language",
	}
	cmd.AddCommand(newExplainTerraformCmd())
	return cmd
}

func newExplainTerraformCmd() *cobra.Command {
	var level string

	cmd := &cobra.Command{
		Use:   "terraform [path]",
		Short: "Explain a Terraform configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			a, err := defaultAdapter(cfg)
			if err != nil {
				return err
			}
			gw := gateway.New(a)
			gw.OnRedact = func(rs []redact.Redaction) {
				fmt.Fprintf(cmd.ErrOrStderr(), "Redacted %d sensitive value(s) before sending to %s.\n", len(rs), a.Name())
			}

			files, err := terraform.ReadTFFiles(dir)
			if err != nil {
				return err
			}
			if len(files) == 0 {
				return errors.New("no .tf files found in " + dir)
			}

			resp, err := gw.Complete(cmd.Context(), provider.Request{
				System: explainSystemPrompt,
				Prompt: fmt.Sprintf("Explain the following Terraform configuration at a %s level.\n\n%s", level, terraform.FilesPrompt(files)),
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), resp.Text)
			return nil
		},
	}
	cmd.Flags().StringVar(&level, "level", "intermediate", "explanation depth: beginner|intermediate|expert")
	return cmd
}
