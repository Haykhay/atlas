package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Haykhay/atlas/internal/config"
	"github.com/Haykhay/atlas/internal/gateway"
	"github.com/Haykhay/atlas/internal/provider"
	"github.com/Haykhay/atlas/internal/redact"
	"github.com/Haykhay/atlas/internal/terraform"
)

func newFixCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Generate pull-request-ready patches for findings (never applies them)",
	}
	cmd.AddCommand(newFixTerraformCmd())
	return cmd
}

func newFixTerraformCmd() *cobra.Command {
	var out string

	cmd := &cobra.Command{
		Use:   "terraform [path]",
		Short: "Generate a unified diff fixing static findings in a Terraform configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}

			resources, err := terraform.ParseDir(dir)
			if err != nil {
				return err
			}
			findings := terraform.RunRules(resources)
			if len(findings) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No findings to fix.")
				return nil
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
			resp, err := gw.Complete(cmd.Context(), provider.Request{
				System: terraform.FixSystemPrompt,
				Prompt: terraform.FixPrompt(files, findings),
			})
			if err != nil {
				return err
			}
			diff, err := terraform.ExtractDiff(resp.Text)
			if err != nil {
				return err
			}

			if out != "" {
				if err := os.WriteFile(out, []byte(diff), 0o644); err != nil { // #nosec G306 -- a patch file is not sensitive
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s — review it, then apply with: git apply %s\n", out, out)
				return nil
			}
			fmt.Fprint(cmd.OutOrStdout(), diff)
			return nil
		},
	}
	cmd.Flags().StringVar(&out, "out", "", "write the patch to a file instead of stdout")
	return cmd
}
