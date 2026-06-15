package cmd

import (
	"context"
	"fmt"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/manager"
	"github.com/lucidfrontier45/i/internal/types"
	"github.com/spf13/cobra"
)

var forceSync bool

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Install all registered packages at their specified versions",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		for name, entry := range cfg.Packages {
			fmt.Printf("syncing %s (%s@%s)...\n", name, entry.Manager, entry.Version)

			drv := manager.Lookup(string(entry.Manager))
			if drv == nil {
				fmt.Printf("error: unknown manager %q\n", entry.Manager)
				hasError = true
				continue
			}

			options := entry.Options
			if forceSync {
				options = copyOptions(options)
				options["force"] = true
			}

			spec := types.PackageSpec{
				Name:     name,
				Version:  entry.Version,
				Manager:  entry.Manager,
				Features: entry.Features,
				Options:  options,
			}

			if err := drv.Install(context.Background(), spec); err != nil {
				fmt.Printf("error syncing %s: %v\n", name, err)
				hasError = true
				continue
			}

			if installedVer, err := drv.InstalledVersion(
				context.Background(),
				string(name),
			); err == nil && installedVer != "" &&
				installedVer != entry.Version {
				entry.Version = installedVer
				cfg.Packages[name] = entry
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
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().
		BoolVarP(&forceSync, "force", "f", false, "Force reinstall even if already installed")
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
