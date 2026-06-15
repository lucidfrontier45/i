package cmd

import (
	"os"

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

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().
		BoolP("verbose", "v", false, "Verbose: show full underlying command of each driver")
}
