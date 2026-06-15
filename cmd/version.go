package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

func runVersion() error {
	fmt.Printf("i version %s\n", Version)
	if Commit != "" {
		fmt.Printf("commit %s\n", Commit)
	}
	if Date != "" {
		fmt.Printf("built at %s\n", Date)
	}
	fmt.Printf("go version %s\n", runtime.Version())
	fmt.Printf("os/arch %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return nil
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return runVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
