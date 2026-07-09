package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/types"
	"github.com/spf13/cobra"
)

func runList() error {
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

	// Precompute manager per package name so the sort comparator avoids
	// repeated map lookups. Output is grouped by manager then alphabetical.
	managers := make(map[string]string, len(cfg.Packages))
	names := make([]string, 0, len(cfg.Packages))
	for name, entry := range cfg.Packages {
		names = append(names, string(name))
		managers[string(name)] = string(entry.Manager)
	}
	sort.Slice(names, func(i, j int) bool {
		if managers[names[i]] != managers[names[j]] {
			return managers[names[i]] < managers[names[j]]
		}
		return names[i] < names[j]
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "Package\tManager\tVersion\tAlias")
	_, _ = fmt.Fprintln(w, "-------\t-------\t-------\t-----")
	for _, name := range names {
		entry := cfg.Packages[types.PackageName(name)]
		version := entry.Version
		if version == "" {
			version = "latest"
		}
		alias := strings.Join(byPkg[name], ", ")
		_, _ = fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\n",
			name,
			string(entry.Manager),
			version,
			alias,
		)
	}
	_ = w.Flush()
	return nil
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered packages",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return runList()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
