package cmd

import (
	"context"
	"fmt"
	"sort"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/manager"
	"github.com/lucidfrontier45/i/internal/types"
	"github.com/spf13/cobra"
)

func runUpgrade(key string) error {
	cfg, _, err := config.Read()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	if len(cfg.Packages) == 0 {
		fmt.Println("no packages registered")
		return nil
	}

	if key != "" {
		full := cfg.ResolveName(key)
		entry, ok := cfg.Packages[full]
		if !ok {
			return fmt.Errorf("package %q not found in config", key)
		}

		drv := manager.Lookup(string(entry.Manager))
		if drv == nil {
			return fmt.Errorf("unknown manager %q", entry.Manager)
		}

		spec := types.PackageSpec{
			Name:     full,
			Version:  entry.Version,
			Manager:  entry.Manager,
			Features: entry.Features,
			Options:  entry.Options,
		}

		fmt.Printf("upgrading %s (%s)...\n", full, entry.Manager)
		if err := drv.Upgrade(context.Background(), spec); err != nil {
			return fmt.Errorf("upgrade %s: %w", full, err)
		}

		if installedVer, err := drv.InstalledVersion(
			context.Background(),
			spec,
		); err == nil && installedVer != "" &&
			installedVer != entry.Version {
			entry.Version = installedVer
			cfg.Packages[full] = entry
			if _, err := config.Write(cfg); err != nil {
				return fmt.Errorf("write config: %w", err)
			}
		}

		return nil
	}

	hasError := false
	needsWrite := false
	names := make([]string, 0, len(cfg.Packages))
	for n := range cfg.Packages {
		names = append(names, string(n))
	}
	sort.Strings(names)
	for _, n := range names {
		name := types.PackageName(n)
		entry := cfg.Packages[name]
		drv := manager.Lookup(string(entry.Manager))
		if drv == nil {
			fmt.Printf("error: unknown manager %q\n", entry.Manager)
			hasError = true
			continue
		}

		spec := types.PackageSpec{
			Name:     name,
			Version:  entry.Version,
			Manager:  entry.Manager,
			Features: entry.Features,
			Options:  entry.Options,
		}

		fmt.Printf("upgrading %s (%s)...\n", name, entry.Manager)
		if err := drv.Upgrade(context.Background(), spec); err != nil {
			fmt.Printf("error upgrading %s: %v\n", name, err)
			hasError = true
			continue
		}

		if installedVer, err := drv.InstalledVersion(
			context.Background(),
			spec,
		); err == nil && installedVer != "" &&
			installedVer != entry.Version {
			entry.Version = installedVer
			cfg.Packages[name] = entry
			needsWrite = true
		}
	}

	if needsWrite {
		if _, err := config.Write(cfg); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
	}

	if hasError {
		return fmt.Errorf("some packages failed to upgrade")
	}

	return nil
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [package]",
	Short: "Upgrade one or all registered packages to the latest version",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		key := ""
		if len(args) == 1 {
			key = args[0]
		}
		return runUpgrade(key)
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
