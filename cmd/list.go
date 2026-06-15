package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
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

		byPkg := make(map[string][]string)
		for alias, full := range cfg.Index {
			byPkg[string(full)] = append(byPkg[string(full)], string(alias))
		}
		for _, aliases := range byPkg {
			sort.Strings(aliases)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		_, _ = fmt.Fprintln(w, "Package\tManager\tVersion\tAlias")
		_, _ = fmt.Fprintln(w, "-------\t-------\t-------\t-----")
		for name, entry := range cfg.Packages {
			version := entry.Version
			if version == "" {
				version = "latest"
			}
			alias := strings.Join(byPkg[string(name)], ", ")
			_, _ = fmt.Fprintf(
				w,
				"%s\t%s\t%s\t%s\n",
				string(name),
				string(entry.Manager),
				version,
				alias,
			)
		}
		_ = w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
