package cmd

import (
	"fmt"

	"github.com/lucidfrontier45/i/internal/config"
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
			pkg := args[0]
			entry, ok := cfg.Packages[pkg]
			if !ok {
				return fmt.Errorf("package %q not found in config", pkg)
			}
			fmt.Printf("upgrading %s (%s)...\n", pkg, entry.Manager)
			// TODO: invoke driver
			_ = entry
			return nil
		}

		hasError := false
		for name, entry := range cfg.Packages {
			fmt.Printf("upgrading %s (%s)...\n", name, entry.Manager)
			// TODO: invoke driver
			_ = name
			_ = entry
		}

		if hasError {
			return fmt.Errorf("some packages failed to upgrade")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(upgradeCmd)
	upgradeCmd.Flags().String("manager", "", "Package manager (bun, uv, cargo)")
}
