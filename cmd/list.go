package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func runList() error {
	cfg, _, err := readConfig()
	if err != nil {
		return err
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

	// Output is grouped by manager then alphabetical.
	names := sortedNamesByManager(cfg.Packages)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "Package\tManager\tVersion\tAlias")
	_, _ = fmt.Fprintln(w, "-------\t-------\t-------\t-----")
	for _, name := range names {
		entry := cfg.Packages[name]
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
