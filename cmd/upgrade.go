package cmd

import (
	"context"
	"fmt"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/types"
	"github.com/spf13/cobra"
)

func showUpgradeResult(
	name types.PackageName,
	oldVersion string,
	spec types.PackageSpec,
	drv types.Driver,
) error {
	if err := recordInstalledVersion(name, drv, spec); err != nil {
		return err
	}
	return printVersionChange(name, oldVersion)
}

// loadEntryAfterUpgrade reloads the config and returns it with name's entry,
// sharing the read-and-require logic between the two post-upgrade steps.
func loadEntryAfterUpgrade(
	name types.PackageName,
) (*config.Config, config.PackageEntry, error) {
	cfg, _, err := config.Read()
	if err != nil {
		return nil, config.PackageEntry{}, fmt.Errorf(
			"read config after upgrading %s: %w",
			name,
			err,
		)
	}
	entry, ok := cfg.Packages[name]
	if !ok {
		return nil, config.PackageEntry{}, fmt.Errorf(
			"package %q disappeared from config after upgrade",
			name,
		)
	}
	return cfg, entry, nil
}

// recordInstalledVersion refreshes the recorded version of name with the version
// the driver reports, if it can be determined.
func recordInstalledVersion(
	name types.PackageName,
	drv types.Driver,
	spec types.PackageSpec,
) error {
	installedVersion, versionErr := drv.InstalledVersion(context.Background(), spec)
	if versionErr != nil {
		fmt.Printf("warning: could not determine installed version of %s: %v\n", name, versionErr)
	}

	cfg, entry, err := loadEntryAfterUpgrade(name)
	if err != nil {
		return err
	}
	if versionErr != nil || installedVersion == "" || installedVersion == entry.Version {
		return nil
	}

	entry.Version = installedVersion
	cfg.Packages[name] = entry
	if _, err := config.Write(cfg); err != nil {
		return fmt.Errorf("write config after upgrading %s: %w", name, err)
	}
	return nil
}

// printVersionChange reports how name's recorded version changed after upgrade.
func printVersionChange(name types.PackageName, oldVersion string) error {
	_, entry, err := loadEntryAfterUpgrade(name)
	if err != nil {
		return err
	}
	if oldVersion == entry.Version {
		fmt.Printf(" no change\n")
		return nil
	}
	fmt.Printf(" %s -> %s\n", oldVersion, entry.Version)
	return nil
}

func runUpgrade(key string) error {
	cfg, _, err := readConfig()
	if err != nil {
		return err
	}

	if len(cfg.Packages) == 0 {
		fmt.Println("no packages registered")
		return nil
	}

	oldVersions := make(map[types.PackageName]string, len(cfg.Packages))
	for name, entry := range cfg.Packages {
		oldVersions[name] = entry.Version
	}

	if key != "" {
		return upgradeOne(cfg, oldVersions, key)
	}
	return upgradeAll(cfg, oldVersions)
}

// upgradeOne upgrades the single package addressed by key.
func upgradeOne(
	cfg *config.Config,
	oldVersions map[types.PackageName]string,
	key string,
) error {
	full, entry, err := requirePackage(cfg, key)
	if err != nil {
		return err
	}

	drv, err := lookupDriver(entry.Manager)
	if err != nil {
		return err
	}

	spec := specFromEntry(full, entry)

	fmt.Printf("upgrading %s (%s)...", full, entry.Manager)
	if err := drv.Upgrade(context.Background(), spec); err != nil {
		return fmt.Errorf("upgrade %s: %w", full, err)
	}

	return showUpgradeResult(full, oldVersions[full], spec, drv)
}

// upgradeAll upgrades every registered package, continuing past failures.
func upgradeAll(
	cfg *config.Config,
	oldVersions map[types.PackageName]string,
) error {
	hasError := false
	for _, name := range sortedNames(cfg.Packages) {
		entry := cfg.Packages[name]

		drv, err := lookupDriver(entry.Manager)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			hasError = true
			continue
		}

		spec := specFromEntry(name, entry)

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
