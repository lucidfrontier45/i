package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/lucidfrontier45/i/internal/selfupdate"
	"github.com/spf13/cobra"
)

func runSelfUpgrade() error {
	fmt.Printf("checking for updates (current: %s)...\n", Version)
	next, err := selfupdate.SelfUpdate(
		context.Background(),
		"lucidfrontier45/i",
		Version,
	)
	if errors.Is(err, selfupdate.ErrUpToDate) {
		fmt.Println("already up to date")
		return nil
	}
	if err != nil {
		return err
	}
	fmt.Printf("upgraded i to %s\n", next)
	fmt.Println("re-run `i version` to confirm")
	return nil
}

var selfUpgradeCmd = &cobra.Command{
	Use:   "self-upgrade",
	Short: "Upgrade the i binary to the latest release",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return runSelfUpgrade()
	},
}

func init() {
	rootCmd.AddCommand(selfUpgradeCmd)
}
