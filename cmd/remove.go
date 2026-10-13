package cmd

import (
	"context"
	"fmt"

	"github.com/lucidfrontier45/i/internal/types"
	"github.com/spf13/cobra"
)

func runRemove(key string) error {
	cfg, path, err := readConfig()
	if err != nil {
		return err
	}

	full, entry, err := requirePackage(cfg, key)
	if err != nil {
		return err
	}

	fmt.Printf("uninstalling %s (%s@%s)...\n", full, entry.Manager, entry.Version)
	drv, err := lookupDriver(entry.Manager)
	if err != nil {
		return err
	}
	spec := types.PackageSpec{
		Name:    full,
		Manager: entry.Manager,
		Options: entry.Options,
	}
	if err := drv.Remove(context.Background(), spec); err != nil {
		return fmt.Errorf("uninstall %s: %w", full, err)
	}

	delete(cfg.Packages, full)
	for a, f := range cfg.Index {
		if f == full {
			delete(cfg.Index, a)
		}
	}

	if _, err := writeConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("removed %s from %s\n", key, path)
	return nil
}

var removeCmd = &cobra.Command{
	Use:   "remove <package>",
	Short: "Uninstall and remove a package from management",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		return runRemove(args[0])
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
