package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/types"
	"github.com/spf13/cobra"
)

// runAliasSet points alias at pkg without touching the installed artifact.
// pkg may be given as the full package name or as an existing alias.
// Re-pointing an alias that already maps elsewhere requires --force.
func runAliasSet(alias, pkgKey string, force bool) error {
	if alias == "" {
		return fmt.Errorf("alias must not be empty")
	}

	cfg, _, err := config.Read()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	aliasKey := types.PackageAlias(alias)
	pkg := cfg.ResolveName(pkgKey)
	if _, ok := cfg.Packages[pkg]; !ok {
		return fmt.Errorf("package %q not found in config", pkgKey)
	}

	if string(aliasKey) == string(pkg) {
		return fmt.Errorf(
			"alias %q is redundant; the package is already addressable by name",
			aliasKey,
		)
	}

	if _, ok := cfg.Packages[types.PackageName(aliasKey)]; ok {
		return fmt.Errorf("alias %q conflicts with an existing package name", aliasKey)
	}

	previous, exists := cfg.Index[aliasKey]
	switch {
	case exists && previous == pkg:
		fmt.Printf("alias %q already maps to %q\n", aliasKey, pkg)
		return nil
	case exists && !force:
		return fmt.Errorf(
			"alias %q already maps to %q; pass --force to re-point it",
			aliasKey,
			previous,
		)
	}

	cfg.Index[aliasKey] = pkg
	path, err := config.Write(cfg)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if exists {
		fmt.Printf("re-pointed alias %s: %q -> %q\n", path, previous, pkg)
	} else {
		fmt.Printf("set alias %s = %q in %s\n", aliasKey, pkg, path)
	}
	return nil
}

// runAliasList prints every alias paired with its full package name.
func runAliasList() error {
	cfg, _, err := config.Read()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	if len(cfg.Index) == 0 {
		fmt.Println("no aliases configured")
		return nil
	}

	aliases := make([]string, 0, len(cfg.Index))
	for alias := range cfg.Index {
		aliases = append(aliases, string(alias))
	}
	sort.Strings(aliases)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "Alias\tPackage")
	_, _ = fmt.Fprintln(w, "-----\t-------")
	for _, alias := range aliases {
		_, _ = fmt.Fprintf(w, "%s\t%s\n", alias, cfg.Index[types.PackageAlias(alias)])
	}
	_ = w.Flush()
	return nil
}

// runAliasRemove deletes a single alias. The argument is keyed strictly by
// alias: a package name is not accepted, and the package and its install are
// left untouched. Dropping every alias for a package is `i remove <pkg>`.
func runAliasRemove(alias string) error {
	cfg, path, err := config.Read()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	aliasKey := types.PackageAlias(alias)
	if _, ok := cfg.Index[aliasKey]; !ok {
		return fmt.Errorf("alias %q not found", aliasKey)
	}

	delete(cfg.Index, aliasKey)
	if _, err := config.Write(cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("removed alias %s from %s\n", aliasKey, path)
	return nil
}

var aliasSetCmd = &cobra.Command{
	Use:   "set <alias> <package>",
	Short: "Point an alias at a registered package",
	Long: "Point an alias at a registered package.\n\n" +
		"<package> may be the full package name or an existing alias.\n" +
		"This edits config only; the package is not reinstalled.",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		force, err := cmd.Flags().GetBool("force")
		if err != nil {
			return err
		}
		return runAliasSet(args[0], args[1], force)
	},
}

var aliasListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured aliases",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return runAliasList()
	},
}

var aliasRemoveCmd = &cobra.Command{
	Use:   "remove <alias>",
	Short: "Delete an alias without touching its package",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runAliasRemove(args[0])
	},
}

var aliasCmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage package aliases",
	Long: "Manage package aliases.\n\n" +
		"Aliases are config-only shortcuts: they change which key addresses a\n" +
		"registered package without reinstalling it.",
}

func init() {
	aliasSetCmd.Flags().
		Bool("force", false, "Re-point an alias that already maps to another package")

	aliasCmd.AddCommand(aliasSetCmd)
	aliasCmd.AddCommand(aliasListCmd)
	aliasCmd.AddCommand(aliasRemoveCmd)
	rootCmd.AddCommand(aliasCmd)
}
