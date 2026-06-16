package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/manager"
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
	Exclude         string
	With            []string
	WithChanged     bool
	FeaturesChanged bool
}

func runAdd(opts AddOptions) error {
	raw := opts.Raw

	pkgStr, features := parseBracket(raw)
	if pkgStr == "" {
		return fmt.Errorf("invalid package name %q", raw)
	}
	pkg := types.PackageName(pkgStr)

	mgrStr := opts.Manager
	if mgrStr == "" {
		return fmt.Errorf("--manager is required")
	}
	mgr := types.ManagerType(mgrStr)

	with := filterEmpty(opts.With)
	if len(with) > 0 && mgr != types.ManagerUv {
		return fmt.Errorf("--with is only valid with --manager uv")
	}

	version := opts.Version
	var aliasFlag types.PackageAlias
	if opts.Alias != "" {
		aliasFlag = types.PackageAlias(opts.Alias)
	}

	options := make(map[string]any)
	if opts.Destination != "" {
		options["destination"] = opts.Destination
	}
	if opts.BinName != "" {
		options["bin-name"] = opts.BinName
	}
	if opts.Exclude != "" {
		options["exclude"] = opts.Exclude
	}

	_, err := config.EnsureDir()
	if err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	cfg, path, err := config.Read()
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	if target, ok := cfg.Index[types.PackageAlias(pkg)]; ok {
		fmt.Printf("alias %s -> %s already registered to %s\n", pkg, target, path)
		return nil
	}
	if aliasFlag != "" {
		if _, ok := cfg.Packages[types.PackageName(aliasFlag)]; ok {
			return fmt.Errorf("alias name %q conflicts with existing package name", aliasFlag)
		}
	}

	drv := manager.Lookup(string(mgr))
	if drv == nil {
		return fmt.Errorf("unknown manager %q", mgr)
	}

	if existing, exists := cfg.Packages[pkg]; exists {
		hasUpdate := aliasFlag != "" ||
			(version != "" && version != existing.Version) ||
			opts.FeaturesChanged || opts.WithChanged || len(options) > 0
		if !hasUpdate {
			fmt.Printf("package %s already registered to %s\n", pkg, path)
			return nil
		}
		if aliasFlag != "" {
			for key, val := range cfg.Index {
				if val == pkg {
					delete(cfg.Index, key)
				}
			}
			cfg.Index[aliasFlag] = pkg
		}

		merged := existing
		changed := false
		if version != "" && version != existing.Version {
			merged.Version = version
			changed = true
		}
		if opts.FeaturesChanged {
			merged.Features = features
			changed = true
		}
		if opts.WithChanged {
			merged.With = with
			changed = true
		}
		if len(options) > 0 {
			merged.Options = options
			changed = true
		}

		if changed {
			spec := types.PackageSpec{
				Name:     pkg,
				Version:  merged.Version,
				Manager:  mgr,
				Features: merged.Features,
				With:     merged.With,
				Options:  merged.Options,
			}
			if err := drv.Install(context.Background(), spec); err != nil {
				return fmt.Errorf("install %s: %w", pkg, err)
			}

			if installedVer, err := drv.InstalledVersion(
				context.Background(),
				string(pkg),
			); err == nil && installedVer != "" {
				merged.Version = installedVer
			}
			cfg.Packages[pkg] = merged
		}

		if _, err := config.Write(cfg); err != nil {
			return fmt.Errorf("write config: %w", err)
		}

		if aliasFlag != "" {
			fmt.Printf("updated alias %s -> %s (manager: %s) to %s\n", aliasFlag, pkg, mgr, path)
		} else {
			fmt.Printf("updated %s (manager: %s) to %s\n", pkg, mgr, path)
		}
		return nil
	}

	spec := types.PackageSpec{
		Name:     pkg,
		Version:  version,
		Manager:  mgr,
		Features: features,
		With:     with,
		Options:  options,
	}

	if err := drv.Install(context.Background(), spec); err != nil {
		return fmt.Errorf("install %s: %w", pkg, err)
	}

	if installedVer, err := drv.InstalledVersion(
		context.Background(),
		string(pkg),
	); err == nil && installedVer != "" {
		version = installedVer
	}

	entry := config.PackageEntry{
		Manager:  mgr,
		Version:  version,
		Features: features,
		With:     with,
	}
	if len(options) > 0 {
		entry.Options = options
	}
	cfg.Packages[pkg] = entry

	display := string(pkg)
	if aliasFlag != "" {
		cfg.Index[aliasFlag] = pkg
		display = string(aliasFlag)
	}

	_, err = config.Write(cfg)
	if err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	fmt.Printf("added %s (manager: %s, version: %s) to %s\n", display, mgr, version, path)
	return nil
}

var addCmd = &cobra.Command{
	Use:   "add <package>",
	Short: "Register a package to manage",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, _ := cmd.Flags().GetString("manager")
		version, _ := cmd.Flags().GetString("version")
		alias, _ := cmd.Flags().GetString("alias")
		dest, _ := cmd.Flags().GetString("destination")
		binName, _ := cmd.Flags().GetString("bin-name")
		exclude, _ := cmd.Flags().GetString("exclude")
		with, _ := cmd.Flags().GetStringSlice("with")
		withChanged := cmd.Flags().Changed("with")
		return runAdd(AddOptions{
			Raw: args[0], Manager: mgr, Version: version,
			Alias: alias, Destination: dest,
			BinName: binName, Exclude: exclude,
			With: with, WithChanged: withChanged,
			FeaturesChanged: strings.Contains(args[0], "["),
		})
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
