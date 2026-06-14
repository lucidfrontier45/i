package cmd

import (
	"context"
	"fmt"

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
		pkg := args[0]

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

		spec := types.PackageSpec{Name: pkg, Version: version, Manager: mgr}

		if err := drv.Install(context.Background(), spec); err != nil {
			return fmt.Errorf("install %s: %w", pkg, err)
		}

		if installedVer, err := drv.InstalledVersion(
			context.Background(),
			pkg,
		); err == nil &&
			installedVer != "" {
			version = installedVer
		}

		cfg.Packages[pkg] = config.PackageEntry{
			Manager: mgr,
			Version: version,
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
