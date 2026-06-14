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

		spec := types.PackageSpec{
			Name:     pkg,
			Version:  version,
			Manager:  mgr,
			Features: features,
		}

		if err := drv.Install(context.Background(), spec); err != nil {
			return fmt.Errorf("install %s: %w", pkg, err)
		}

		lookupPkg := pkg
		if len(features) > 0 {
			lookupPkg = pkg + "[" + strings.Join(features, ",") + "]"
		}
		if installedVer, err := drv.InstalledVersion(
			context.Background(),
			lookupPkg,
		); err == nil &&
			installedVer != "" {
			version = installedVer
		}

		cfg.Packages[pkg] = config.PackageEntry{
			Manager:  mgr,
			Version:  version,
			Features: features,
		}

		_, err = config.Write(cfg)
		if err != nil {
			return fmt.Errorf("write config: %w", err)
		}

		fmt.Printf("added %s (manager: %s, version: %s) to %s\n", pkg, mgr, version, path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.Flags().String("manager", "", "Package manager (bun, uv, cargo)")
	_ = addCmd.MarkFlagRequired("manager")
	addCmd.Flags().String("version", "", "Version to install")
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
