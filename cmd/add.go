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

var addCmd = &cobra.Command{
	Use:   "add <package>",
	Short: "Register a package to manage",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		raw := args[0]

		pkg, features := parseBracket(raw)
		if pkg == "" {
			return fmt.Errorf("invalid package name %q", raw)
		}

		mgr, _ := cmd.Flags().GetString("manager")
		if mgr == "" {
			return fmt.Errorf("--manager is required")
		}

		version, _ := cmd.Flags().GetString("version")
		aliasFlag, _ := cmd.Flags().GetString("alias")
		destination, _ := cmd.Flags().GetString("destination")
		binName, _ := cmd.Flags().GetString("bin-name")
		exclude, _ := cmd.Flags().GetString("exclude")

		options := make(map[string]any)
		if destination != "" {
			options["destination"] = destination
		}
		if binName != "" {
			options["bin-name"] = binName
		}
		if exclude != "" {
			options["exclude"] = exclude
		}

		_, err := config.EnsureDir()
		if err != nil {
			return fmt.Errorf("create config dir: %w", err)
		}

		cfg, path, err := config.Read()
		if err != nil {
			return fmt.Errorf("read config: %w", err)
		}

		drv := manager.Lookup(mgr)
		if drv == nil {
			return fmt.Errorf("unknown manager %q", mgr)
		}

		if existing, exists := cfg.Packages[pkg]; exists {
			if aliasFlag == "" {
				return fmt.Errorf("package %q already registered", pkg)
			}

			cfg.Index[aliasFlag] = pkg

			if version != "" && version != existing.Version {
				spec := types.PackageSpec{
					Name:     pkg,
					Version:  version,
					Manager:  mgr,
					Features: features,
					Options:  options,
				}
				if err := drv.Install(context.Background(), spec); err != nil {
					return fmt.Errorf("install %s: %w", pkg, err)
				}

				if installedVer, err := drv.InstalledVersion(
					context.Background(),
					pkg,
				); err == nil && installedVer != "" {
					version = installedVer
				}

				existing.Version = version
				existing.Features = features
				if len(options) > 0 {
					existing.Options = options
				}
				cfg.Packages[pkg] = existing
			}

			if _, err := config.Write(cfg); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			fmt.Printf("added alias %s -> %s (manager: %s) to %s\n", aliasFlag, pkg, mgr, path)
			return nil
		}

		spec := types.PackageSpec{
			Name:     pkg,
			Version:  version,
			Manager:  mgr,
			Features: features,
			Options:  options,
		}

		if err := drv.Install(context.Background(), spec); err != nil {
			return fmt.Errorf("install %s: %w", pkg, err)
		}

		if installedVer, err := drv.InstalledVersion(
			context.Background(),
			pkg,
		); err == nil && installedVer != "" {
			version = installedVer
		}

		entry := config.PackageEntry{
			Manager:  mgr,
			Version:  version,
			Features: features,
		}
		if len(options) > 0 {
			entry.Options = options
		}
		cfg.Packages[pkg] = entry

		display := pkg
		if aliasFlag != "" {
			cfg.Index[aliasFlag] = pkg
			display = aliasFlag
		}

		_, err = config.Write(cfg)
		if err != nil {
			return fmt.Errorf("write config: %w", err)
		}

		fmt.Printf("added %s (manager: %s, version: %s) to %s\n", display, mgr, version, path)
		return nil
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
