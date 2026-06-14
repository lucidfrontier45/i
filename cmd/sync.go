package cmd

import (
	"fmt"

	"github.com/lucidfrontier45/i/internal/config"
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
		for name, entry := range cfg.Packages {
			fmt.Printf("syncing %s (%s@%s)...\n", name, entry.Manager, entry.Version)
			// TODO: invoke driver
			_ = name
			_ = entry
			_ = forceSync
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
