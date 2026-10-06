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

func showUpgradeResult(
	name types.PackageName,
	oldVersion string,
	spec types.PackageSpec,
	drv types.Driver,
) error {
	installedVersion, versionErr := drv.InstalledVersion(context.Background(), spec)
	if versionErr != nil {
		fmt.Printf("warning: could not determine installed version of %s: %v\n", name, versionErr)
	}

	cfg, _, err := config.Read()
	if err != nil {
		return fmt.Errorf("read config after upgrading %s: %w", name, err)
	}
	entry, ok := cfg.Packages[name]
	if !ok {
		return fmt.Errorf("package %q disappeared from config after upgrade", name)
	}
	if versionErr == nil && installedVersion != "" && installedVersion != entry.Version {
		entry.Version = installedVersion
		cfg.Packages[name] = entry
		if _, err := config.Write(cfg); err != nil {
			return fmt.Errorf("write config after upgrading %s: %w", name, err)
		}
	}

	cfg, _, err = config.Read()
	if err != nil {
		return fmt.Errorf("read config after upgrading %s: %w", name, err)
	}
	latestVersion, ok := cfg.Packages[name]
	if !ok {
		return fmt.Errorf("package %q disappeared from config after upgrade", name)
	}
	if oldVersion == latestVersion.Version {
		fmt.Printf(" no change\n")
	} else {
		fmt.Printf(" %s -> %s\n", oldVersion, latestVersion.Version)
	}
	return nil
}

func runUpgrade(key string) error {
	cfg, _, err := config.Read()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}
	oldVersions := make(map[types.PackageName]string, len(cfg.Packages))
	for name, entry := range cfg.Packages {
		oldVersions[name] = entry.Version
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

		fmt.Printf("upgrading %s (%s)...", full, entry.Manager)
		if err := drv.Upgrade(context.Background(), spec); err != nil {
			return fmt.Errorf("upgrade %s: %w", full, err)
		}

		if err := showUpgradeResult(full, oldVersions[full], spec, drv); err != nil {
			return err
		}
		return nil
	}

	hasError := false
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

		fmt.Printf("upgrading %s (%s)...", name, entry.Manager)
		if err := drv.Upgrade(context.Background(), spec); err != nil {
			fmt.Printf("error upgrading %s: %v\n", name, err)
			hasError = true
			continue
		}

		if err := showUpgradeResult(name, oldVersions[name], spec, drv); err != nil {
			fmt.Printf("error recording upgrade result for %s: %v\n", name, err)
			hasError = true
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
