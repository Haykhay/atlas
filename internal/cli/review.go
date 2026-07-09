package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/Haykhay/atlas/internal/config"
	"github.com/Haykhay/atlas/internal/finding"
	"github.com/Haykhay/atlas/internal/gateway"
	"github.com/Haykhay/atlas/internal/redact"
	"github.com/Haykhay/atlas/internal/review"
	"github.com/Haykhay/atlas/internal/waf"
)

func newReviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "review",
		Short: "Review infrastructure for risks and Well-Architected alignment",
	}
	cmd.AddCommand(newReviewTerraformCmd())
	return cmd
}

func newReviewTerraformCmd() *cobra.Command {
	var format string
	var offline bool

	cmd := &cobra.Command{
		Use:   "terraform [path]",
		Short: "Review Terraform against the AWS Well-Architected Framework",
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

			r := &review.Terraform{}
			if !offline {
				a, err := defaultAdapter(cfg)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "Note: static analysis only (%v)\n", err)
				} else {
					gw := gateway.New(a)
					gw.OnRedact = func(rs []redact.Redaction) {
						fmt.Fprintf(cmd.ErrOrStderr(), "Redacted %d sensitive value(s) before sending to %s.\n", len(rs), a.Name())
					}
					r.AI = gw
				}
			}

			report, err := r.Run(cmd.Context(), dir)
			if err != nil {
				return err
			}
			for _, w := range report.Warnings {
				fmt.Fprintln(cmd.ErrOrStderr(), "Warning:", w)
			}

			if format == "json" {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(report)
			}
			renderScores(cmd.OutOrStdout(), report.PillarScores)
			return finding.RenderHuman(cmd.OutOrStdout(), report.Findings)
		},
	}
	cmd.Flags().StringVar(&format, "format", "human", "output format: human|json")
	cmd.Flags().BoolVar(&offline, "offline", false, "static analysis only; never contact an AI provider")
	return cmd
}

func renderScores(w io.Writer, scores []waf.PillarScore) {
	fmt.Fprintln(w, "AWS Well-Architected Scores")
	for _, s := range scores {
		fmt.Fprintf(w, "  %-26s %3d/100  (%d finding(s))\n", s.Pillar, s.Score, s.Findings)
	}
	fmt.Fprintln(w)
}
