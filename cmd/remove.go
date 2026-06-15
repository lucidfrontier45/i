package cmd

import (
	"context"
	"fmt"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/manager"
	"github.com/lucidfrontier45/i/internal/types"
	"github.com/spf13/cobra"
)

func runRemove(key string) error {
	cfg, path, err := config.Read()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	full := cfg.ResolveName(key)
	entry, ok := cfg.Packages[full]
	if !ok {
		return fmt.Errorf("package %q not found in config", key)
	}

	fmt.Printf("uninstalling %s (%s@%s)...\n", full, entry.Manager, entry.Version)
	drv := manager.Lookup(string(entry.Manager))
	if drv == nil {
		return fmt.Errorf("unknown manager %q", entry.Manager)
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

	_, err = config.Write(cfg)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
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
