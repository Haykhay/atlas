package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/silverstreaminnovations/atlas/internal/config"
	"github.com/silverstreaminnovations/atlas/internal/provider"
)

func newConfigureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Select and authenticate an AI provider",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			names := provider.Names()
			fmt.Fprintln(cmd.OutOrStdout(), "Select an AI provider:")
			for i, name := range names {
				fmt.Fprintf(cmd.OutOrStdout(), "  %d) %s\n", i+1, name)
			}
			fmt.Fprint(cmd.OutOrStdout(), "Choice: ")
			line, err := readLine(cmd)
			if err != nil {
				return err
			}
			idx, err := strconv.Atoi(line)
			if err != nil || idx < 1 || idx > len(names) {
				return fmt.Errorf("invalid selection %q: enter a number between 1 and %d", line, len(names))
			}
			name := names[idx-1]
			if err := loginProvider(cmd, cfg, name); err != nil {
				return err
			}
			cfg.DefaultProvider = name
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Done. %s is your default provider.\n", name)
			return nil
		},
	}
}
