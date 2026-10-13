package cmd

import (
	"context"
	"errors"
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

func runUpgrade(keys []string) error {
	cfg, names, ok, err := loadTargets(keys)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	oldVersions := make(map[types.PackageName]string, len(cfg.Packages))
	for name, entry := range cfg.Packages {
		oldVersions[name] = entry.Version
	}

	var errs []error
	for _, name := range names {
		if err := upgradePackage(cfg, name, oldVersions[name]); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// upgradePackage upgrades a single registered package to its latest version and
// records the resulting version in the config. Failures are returned wrapped
// with the package name so the caller can report them together.
func upgradePackage(
	cfg *config.Config,
	name types.PackageName,
	oldVersion string,
) error {
	entry := cfg.Packages[name]

	drv, err := lookupDriver(entry.Manager)
	if err != nil {
		return fmt.Errorf("upgrade %s: %w", name, err)
	}

	spec := specFromEntry(name, entry)

	fmt.Printf("upgrading %s (%s)...", name, entry.Manager)
	if err := drv.Upgrade(context.Background(), spec); err != nil {
		fmt.Println()
		return fmt.Errorf("upgrade %s: %w", name, err)
	}

	return showUpgradeResult(name, oldVersion, spec, drv)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade [package...]",
	Short: "Upgrade registered packages to the latest version",
	Args:  cobra.ArbitraryArgs,
	RunE: func(_ *cobra.Command, args []string) error {
		return runUpgrade(args)
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
}
