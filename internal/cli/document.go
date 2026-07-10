package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Haykhay/atlas/internal/config"
	"github.com/Haykhay/atlas/internal/docgen"
	"github.com/Haykhay/atlas/internal/gateway"
	"github.com/Haykhay/atlas/internal/provider"
	"github.com/Haykhay/atlas/internal/redact"
	"github.com/Haykhay/atlas/internal/terraform"
)

func newDocumentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "document",
		Short: "Generate infrastructure documentation",
	}
	cmd.AddCommand(newDocumentTerraformCmd())
	return cmd
}

func newDocumentTerraformCmd() *cobra.Command {
	var docType string
	var out string

	cmd := &cobra.Command{
		Use:   "terraform [path]",
		Short: "Generate documentation for a Terraform configuration",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
			}
			switch docType {
			case "architecture", "runbook", "adr", "readme":
			default:
				return fmt.Errorf("unknown --type %q (architecture|runbook|adr|readme)", docType)
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}

			var doc strings.Builder

			// The architecture skeleton is static — it always works.
			if docType == "architecture" {
				resources, err := terraform.ParseDir(dir)
				if err != nil {
					return err
				}
				doc.WriteString(docgen.Skeleton("Architecture — "+dir, resources))
			}

			adapter, adapterErr := defaultAdapter(cfg)
			if adapterErr != nil {
				if docType != "architecture" {
					return adapterErr // prose-only doc types need a provider
				}
				fmt.Fprintf(cmd.ErrOrStderr(), "Note: skeleton only, AI prose skipped (%v)\n", adapterErr)
			} else {
				gw := gateway.New(adapter)
				gw.OnRedact = func(rs []redact.Redaction) {
					fmt.Fprintf(cmd.ErrOrStderr(), "Redacted %d sensitive value(s) before sending to %s.\n", len(rs), adapter.Name())
				}
				files, err := terraform.ReadTFFiles(dir)
				if err != nil {
					return err
				}
				if len(files) == 0 {
					return fmt.Errorf("no .tf files found in %s", dir)
				}
				resp, err := gw.Complete(cmd.Context(), provider.Request{
					System: docgen.SystemPrompt(docType),
					Prompt: docgen.FilesPrompt(files),
				})
				if err != nil {
					if docType != "architecture" {
						return err
					}
					fmt.Fprintln(cmd.ErrOrStderr(), "Warning: AI prose unavailable:", err)
				} else {
					if doc.Len() > 0 {
						doc.WriteString("\n")
					}
					doc.WriteString(resp.Text)
					doc.WriteString("\n")
				}
			}

			if out != "" {
				if err := os.WriteFile(out, []byte(doc.String()), 0o644); err != nil { // #nosec G306 -- documentation is not sensitive
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", out)
				return nil
			}
			fmt.Fprint(cmd.OutOrStdout(), doc.String())
			return nil
		},
	}
	cmd.Flags().StringVar(&docType, "type", "architecture", "document type: architecture|runbook|adr|readme")
	cmd.Flags().StringVar(&out, "out", "", "write the document to a file instead of stdout")
	return cmd
}
