package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/types"
	"github.com/spf13/cobra"
)

type AddOptions struct {
	Raw             string
	Manager         string
	Version         string
	Alias           string
	Destination     string
	BinName         string
	Rename          string
	Exclude         string
	FeaturesChanged bool
}

// addUpdate carries the resolved inputs shared by the existing-package and
// new-package paths of `i add`.
type addUpdate struct {
	pkg             types.PackageName
	mgr             types.ManagerType
	version         string
	features        []string
	options         map[string]any
	aliasFlag       types.PackageAlias
	featuresChanged bool
	withChanged     bool
}

func runAdd(opts AddOptions, with []string, withChanged bool) error {
	up, err := parseAddRequest(opts, with, withChanged)
	if err != nil {
		return err
	}

	if _, err := config.EnsureDir(); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	cfg, path, err := readConfig()
	if err != nil {
		return err
	}

	if up.aliasFlag != "" {
		if err := checkAliasConflict(cfg, up.aliasFlag, up.pkg); err != nil {
			return err
		}
	}

	drv, err := lookupDriver(up.mgr)
	if err != nil {
		return err
	}

	if _, exists := cfg.Packages[up.pkg]; exists {
		return updateExistingPackage(cfg, path, drv, up)
	}
	return createNewPackage(cfg, path, drv, up)
}

// parseAddRequest validates the CLI inputs and resolves them into an addUpdate.
func parseAddRequest(opts AddOptions, with []string, withChanged bool) (addUpdate, error) {
	pkgStr, features := parseBracket(opts.Raw)
	if pkgStr == "" {
		return addUpdate{}, fmt.Errorf("invalid package name %q", opts.Raw)
	}
	if opts.Manager == "" {
		return addUpdate{}, fmt.Errorf("--manager is required")
	}
	mgr := types.ManagerType(opts.Manager)

	with = filterEmpty(with)
	if len(with) > 0 && mgr != types.ManagerUv {
		return addUpdate{}, fmt.Errorf("--with is only valid with --manager uv")
	}

	return addUpdate{
		pkg:             types.PackageName(pkgStr),
		mgr:             mgr,
		version:         opts.Version,
		features:        features,
		options:         buildAddOptions(opts, with),
		aliasFlag:       types.PackageAlias(opts.Alias),
		featuresChanged: opts.FeaturesChanged,
		withChanged:     withChanged,
	}, nil
}

// buildAddOptions assembles the driver-specific option map from the CLI flags.
func buildAddOptions(opts AddOptions, with []string) map[string]any {
	options := make(map[string]any)
	if opts.Destination != "" {
		options["destination"] = opts.Destination
	}
	if opts.BinName != "" {
		options["bin-name"] = opts.BinName
	}
	if opts.Rename != "" {
		options["rename"] = opts.Rename
	}
	if opts.Exclude != "" {
		options["exclude"] = opts.Exclude
	}
	if len(with) > 0 {
		options["with"] = with
	}
	return options
}

// checkAliasConflict rejects an alias that shadows a package name or is already
// bound to a different package.
func checkAliasConflict(
	cfg *config.Config,
	aliasFlag types.PackageAlias,
	pkg types.PackageName,
) error {
	if _, ok := cfg.Packages[types.PackageName(aliasFlag)]; ok {
		return fmt.Errorf("alias %q conflicts with an existing package name", aliasFlag)
	}
	if existing, ok := cfg.Index[aliasFlag]; ok && existing != pkg {
		return fmt.Errorf("alias %q already maps to %q", aliasFlag, existing)
	}
	return nil
}

// hasUpdate reports whether this invocation changes the existing entry at all.
func (up addUpdate) hasUpdate(existing config.PackageEntry) bool {
	return up.aliasFlag != "" ||
		(up.version != "" && up.version != existing.Version) ||
		up.featuresChanged || up.withChanged || len(up.options) > 0
}

// mergeEntry folds the invocation's version, features, and options into an
// existing entry, returning whether anything changed. Option keys are merged
// per-key: existing keys not mentioned here are preserved.
func mergeEntry(existing config.PackageEntry, up addUpdate) (config.PackageEntry, bool) {
	merged := existing
	changed := false
	if up.version != "" && up.version != existing.Version {
		merged.Version = up.version
		changed = true
	}
	if up.featuresChanged {
		merged.Features = up.features
		changed = true
	}
	if len(up.options) > 0 {
		if merged.Options == nil {
			merged.Options = make(map[string]any)
		}
		for k, v := range up.options {
			merged.Options[k] = v
		}
		changed = true
	}
	return merged, changed
}

// installAndQueryVersion installs spec and returns the version the driver
// reports afterwards, falling back to fallback when it cannot be determined.
func installAndQueryVersion(
	drv types.Driver,
	spec types.PackageSpec,
	fallback string,
) (string, error) {
	ctx := context.Background()
	if err := drv.Install(ctx, spec); err != nil {
		return "", err
	}
	if installed, err := drv.InstalledVersion(ctx, spec); err == nil && installed != "" {
		return installed, nil
	}
	return fallback, nil
}

// updateExistingPackage applies an update to an already-registered package.
func updateExistingPackage(
	cfg *config.Config,
	path string,
	drv types.Driver,
	up addUpdate,
) error {
	existing := cfg.Packages[up.pkg]
	if up.mgr != existing.Manager {
		return fmt.Errorf(
			"package %q is managed by %q; run 'i remove %s' first to switch managers",
			up.pkg,
			existing.Manager,
			up.pkg,
		)
	}

	if !up.hasUpdate(existing) {
		fmt.Printf("package %s already registered to %s\n", up.pkg, path)
		return nil
	}

	// Additive: prior aliases pointing at pkg are kept. Many aliases may
	// point at one package; `i alias remove` drops a single one.
	if up.aliasFlag != "" {
		cfg.Index[up.aliasFlag] = up.pkg
	}

	merged, changed := mergeEntry(existing, up)
	if changed {
		spec := specFromEntry(up.pkg, merged)
		installed, err := installAndQueryVersion(drv, spec, merged.Version)
		if err != nil {
			return fmt.Errorf("install %s: %w", up.pkg, err)
		}
		merged.Version = installed
		cfg.Packages[up.pkg] = merged
	}

	if _, err := writeConfig(cfg); err != nil {
		return err
	}

	printAddResult(up, path)
	return nil
}

// createNewPackage installs and registers a package that is not yet known.
func createNewPackage(
	cfg *config.Config,
	path string,
	drv types.Driver,
	up addUpdate,
) error {
	entry := config.PackageEntry{
		Manager:  up.mgr,
		Version:  up.version,
		Features: up.features,
	}
	if len(up.options) > 0 {
		entry.Options = up.options
	}

	spec := specFromEntry(up.pkg, entry)
	installed, err := installAndQueryVersion(drv, spec, up.version)
	if err != nil {
		return fmt.Errorf("install %s: %w", up.pkg, err)
	}
	entry.Version = installed
	cfg.Packages[up.pkg] = entry

	display := string(up.pkg)
	if up.aliasFlag != "" {
		cfg.Index[up.aliasFlag] = up.pkg
		display = string(up.aliasFlag)
	}

	if _, err := writeConfig(cfg); err != nil {
		return err
	}

	fmt.Printf("added %s (manager: %s, version: %s) to %s\n", display, up.mgr, entry.Version, path)
	return nil
}

// printAddResult reports the outcome of updating an existing package.
func printAddResult(up addUpdate, path string) {
	if up.aliasFlag != "" {
		fmt.Printf(
			"updated alias %s -> %s (manager: %s) to %s\n",
			up.aliasFlag,
			up.pkg,
			up.mgr,
			path,
		)
		return
	}
	fmt.Printf("updated %s (manager: %s) to %s\n", up.pkg, up.mgr, path)
}

var addCmd = &cobra.Command{
	Use:   "add <package>",
	Short: "Register a package to manage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, _ := cmd.Flags().GetString("manager")
		version, _ := cmd.Flags().GetString("version")
		alias, _ := cmd.Flags().GetString("alias")
		dest, _ := cmd.Flags().GetString("destination")
		binName, _ := cmd.Flags().GetString("bin-name")
		rename, _ := cmd.Flags().GetString("rename")
		exclude, _ := cmd.Flags().GetString("exclude")
		with, _ := cmd.Flags().GetStringSlice("with")
		withChanged := cmd.Flags().Changed("with")
		return runAdd(AddOptions{
			Raw: args[0], Manager: mgr, Version: version,
			Alias: alias, Destination: dest,
			BinName: binName, Rename: rename, Exclude: exclude,
			FeaturesChanged: strings.Contains(args[0], "["),
		}, with, withChanged)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().String("manager", "", "Package manager (bun, uv, cargo, grd, go, npm)")
	_ = addCmd.MarkFlagRequired("manager")
	addCmd.Flags().String("version", "", "Version to install")
	addCmd.Flags().
		StringP("alias", "a", "", "Alias name to register the package under (defaults to the package name)")
	addCmd.Flags().String("destination", "", "Destination directory (grd)")
	addCmd.Flags().String("bin-name", "", "Override binary name (grd)")
	addCmd.Flags().String("rename", "", "Rename the installed binary (grd)")
	addCmd.Flags().String("exclude", "", "Comma-separated asset-name substrings to exclude (grd)")
	addCmd.Flags().
		StringSlice("with", nil, "Extra package(s) to include (uv only); repeatable or comma-separated, accepts name or name==version")
}

func parseBracket(raw string) (name string, features []string) {
	idx := strings.IndexByte(raw, '[')
	if idx == -1 {
		return raw, nil
	}
	name = raw[:idx]
	inner := raw[idx+1:]
	if end := strings.IndexByte(inner, ']'); end != -1 {
		inner = inner[:end]
	}
	for _, f := range strings.Split(inner, ",") {
		f = strings.TrimSpace(f)
		if f != "" {
			features = append(features, f)
		}
	}
	return name, features
}

// filterEmpty drops blank entries from a slice and returns a fresh slice.
func filterEmpty(in []string) []string {
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
