package cmd

import (
	"context"
	"fmt"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/manager"
	"github.com/lucidfrontier45/i/internal/types"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [package]",
	Short: "Upgrade one or all registered packages to the latest version",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, _, err := config.Read()
		if err != nil {
			return fmt.Errorf("read config: %w", err)
		}

		if len(cfg.Packages) == 0 {
			fmt.Println("no packages registered")
			return nil
		}

		if len(args) == 1 {
			key := args[0]
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
				string(full),
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
		for name, entry := range cfg.Packages {
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
				string(name),
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
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().String("manager", "", "Package manager (bun, uv, cargo, grd, go, npm)")
}
