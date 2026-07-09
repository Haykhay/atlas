package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Haykhay/atlas/internal/aireview"
	"github.com/Haykhay/atlas/internal/config"
	"github.com/Haykhay/atlas/internal/gateway"
	"github.com/Haykhay/atlas/internal/kubernetes"
	"github.com/Haykhay/atlas/internal/provider"
	"github.com/Haykhay/atlas/internal/redact"
	"github.com/Haykhay/atlas/internal/terraform"
)

const explainSystemPrompt = `You are Atlas, an infrastructure engineering assistant. Explain infrastructure configurations accurately using correct terminology. Structure the explanation by resource, describe how the pieces connect, and call out anything surprising. Do not invent resources that are not in the configuration.`

func newExplainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain",
		Short: "Explain infrastructure in plain language",
	}
	cmd.AddCommand(
		newExplainSubCmd("terraform", "Explain a Terraform configuration", "Terraform configuration", terraform.ReadTFFiles),
		newExplainSubCmd("kubernetes", "Explain Kubernetes manifests", "Kubernetes manifests", kubernetes.ReadManifests),
	)
	return cmd
}

// newExplainSubCmd builds one domain's explain subcommand; all domains
// share this code path.
func newExplainSubCmd(domain, short, label string, read func(string) (map[string]string, error)) *cobra.Command {
	var level string

	cmd := &cobra.Command{
		Use:   domain + " [path]",
		Short: short,
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

			files, err := read(dir)
			if err != nil {
				return err
			}
			if len(files) == 0 {
				return errors.New("no " + domain + " files found in " + dir)
			}

			resp, err := gw.Complete(cmd.Context(), provider.Request{
				System: explainSystemPrompt,
				Prompt: fmt.Sprintf("Explain the following %s at a %s level.\n\n%s",
					label, level, aireview.FilesPrompt(label+":", files)),
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
