package cmd

import (
	"os"
	"path/filepath"

	"github.com/lucidfrontier45/i/internal/manager"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "i",
	Short: "Global Installer Manager",
	Long:  "Uniform interface to many package managers with global install features.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		manager.Verbose = verbose
	},
}

// cleanupToRemove removes a leftover .to_remove file from a previous Windows self-update.
func cleanupToRemove() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	_ = os.Remove(exe + ".to_remove")
}

func Execute() {
	cleanupToRemove()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().
		BoolP("verbose", "v", false, "Verbose: show full underlying command of each driver")
}
