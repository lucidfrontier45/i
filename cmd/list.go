package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered packages",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _, err := config.Read()
		if err != nil {
			return fmt.Errorf("read config: %w", err)
		}

		if len(cfg.Packages) == 0 {
			fmt.Println("no packages registered")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "Package\tManager\tVersion")
		_, _ = fmt.Fprintln(w, "-------\t-------\t-------")
		for name, entry := range cfg.Packages {
			version := entry.Version
			if version == "" {
				version = "latest"
			}
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", name, entry.Manager, version)
		}
		_ = w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
