package cmd

import (
	"context"
	"fmt"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/manager"
	"github.com/lucidfrontier45/i/internal/types"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <package>",
	Short: "Uninstall and remove a package from management",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pkg := args[0]

		cfg, path, err := config.Read()
		if err != nil {
			return fmt.Errorf("read config: %w", err)
		}

		entry, ok := cfg.Packages[pkg]
		if !ok {
			return fmt.Errorf("package %q not found in config", pkg)
		}

		fmt.Printf("uninstalling %s (%s@%s)...\n", pkg, entry.Manager, entry.Version)
		drv := manager.Lookup(entry.Manager)
		if drv == nil {
			return fmt.Errorf("unknown manager %q", entry.Manager)
		}
		spec := types.PackageSpec{
			Name:    pkg,
			Manager: entry.Manager,
		}
		if err := drv.Remove(context.Background(), spec); err != nil {
			return fmt.Errorf("uninstall %s: %w", pkg, err)
		}

		delete(cfg.Packages, pkg)

		_, err = config.Write(cfg)
		if err != nil {
			return fmt.Errorf("write config: %w", err)
		}

		fmt.Printf("removed %s from %s\n", pkg, path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
