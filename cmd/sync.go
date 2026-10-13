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

func runSync(force bool) error {
	cfg, _, err := config.Read()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	if len(cfg.Packages) == 0 {
		fmt.Println("no packages registered")
		return nil
	}

	hasError := false
	needsWrite := false
	for _, name := range sortedNames(cfg.Packages) {
		changed, err := syncPackage(cfg, name, force)
		if err != nil {
			hasError = true
			continue
		}
		if changed {
			needsWrite = true
		}
	}

	if needsWrite {
		if _, err := config.Write(cfg); err != nil {
			return fmt.Errorf("write config: %w", err)
		}
	}

	if hasError {
		return fmt.Errorf("some packages failed to sync")
	}

	return nil
}

// syncPackage installs a single registered package and refreshes its recorded
// version, returning whether the config entry changed. A non-nil error means the
// package failed; the reason is already printed to the user.
func syncPackage(
	cfg *config.Config,
	name types.PackageName,
	force bool,
) (bool, error) {
	entry := cfg.Packages[name]
	fmt.Printf("syncing %s (%s@%s)...\n", name, entry.Manager, entry.Version)

	drv := manager.Lookup(string(entry.Manager))
	if drv == nil {
		fmt.Printf("error: unknown manager %q\n", entry.Manager)
		return false, fmt.Errorf("unknown manager %q", entry.Manager)
	}

	spec := specFromEntry(name, entry)
	if force {
		spec.Options = copyOptions(spec.Options)
		spec.Options["force"] = true
	}

	if err := drv.Install(context.Background(), spec); err != nil {
		fmt.Printf("error syncing %s: %v\n", name, err)
		return false, err
	}

	updated, ok := refreshInstalledVersion(drv, spec, entry)
	if !ok {
		return false, nil
	}
	cfg.Packages[name] = updated
	return true, nil
}

// refreshInstalledVersion re-queries the driver for the actually installed
// version and reports the updated entry when it differs from the recorded one.
func refreshInstalledVersion(
	drv types.Driver,
	spec types.PackageSpec,
	entry config.PackageEntry,
) (config.PackageEntry, bool) {
	installedVer, err := drv.InstalledVersion(context.Background(), spec)
	if err != nil || installedVer == "" || installedVer == entry.Version {
		return entry, false
	}
	entry.Version = installedVer
	return entry, true
}

// sortedNames returns the registered package names in ascending order.
func sortedNames(pkgs map[types.PackageName]config.PackageEntry) []types.PackageName {
	names := make([]types.PackageName, 0, len(pkgs))
	for n := range pkgs {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })
	return names
}

// specFromEntry builds the install spec for a registered package entry.
func specFromEntry(name types.PackageName, entry config.PackageEntry) types.PackageSpec {
	return types.PackageSpec{
		Name:     name,
		Version:  entry.Version,
		Manager:  entry.Manager,
		Features: entry.Features,
		Options:  entry.Options,
	}
}

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Install all registered packages at their specified versions",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, _ []string) error {
		force, _ := cmd.Flags().GetBool("force")
		return runSync(force)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().
		BoolP("force", "f", false, "Force reinstall even if already installed")
}

// copyOptions returns a shallow copy of m so callers can mutate it without
// aliasing the original map (e.g. when injecting `force` during sync).
func copyOptions(m map[string]any) map[string]any {
	if m == nil {
		return make(map[string]any)
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
