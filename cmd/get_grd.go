package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lucidfrontier45/i/internal/selfupdate"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

var (
	getGrdVer    string
	getGrdLatest bool
)

func runGetGrd() error {
	ctx := context.Background()
	version := getGrdVer

	if getGrdLatest {
		rel, err := selfupdate.LatestRelease(ctx, "lucidfrontier45/grd")
		if err != nil {
			return fmt.Errorf("fetch latest release: %w", err)
		}
		version = rel.TagName
	}

	grdPath, err := exec.LookPath("grd")
	onPath := err == nil

	if !onPath {
		iExe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("locate i executable: %w", err)
		}
		dest := filepath.Dir(iExe)
		return downloadGrd(ctx, dest, version)
	}

	currentVer, err := getGrdVersion()
	if err != nil {
		return fmt.Errorf("get grd version: %w", err)
	}

	if semver.Compare(normalizeV(currentVer), normalizeV(version)) >= 0 {
		fmt.Printf("grd %s is already up to date\n", currentVer)
		return nil
	}

	dest := filepath.Dir(grdPath)
	return downloadGrd(ctx, dest, version)
}

func downloadGrd(ctx context.Context, dest, version string) error {
	target, ok := goosArchToRustTarget(runtime.GOOS, runtime.GOARCH)
	if !ok {
		return fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	assetName := "grd-" + target
	if runtime.GOOS == "windows" {
		assetName += ".exe"
	}

	rel, err := selfupdate.ReleaseByTag(ctx, "lucidfrontier45/grd", version)
	if err != nil {
		return fmt.Errorf("fetch release %s: %w", version, err)
	}

	downloadURL, ok := selfupdate.FindAsset(rel.Assets, assetName)
	if !ok {
		return fmt.Errorf(
			"no prebuilt binary for %s/%s (wanted asset %q)",
			runtime.GOOS,
			runtime.GOARCH,
			assetName,
		)
	}

	fmt.Printf("downloading %s...\n", downloadURL)
	data, err := selfupdate.DownloadURL(ctx, downloadURL)
	if err != nil {
		return fmt.Errorf("download %s: %w", assetName, err)
	}

	binName := "grd"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	outPath := filepath.Join(dest, binName)

	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", outPath, err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(outPath, 0o755); err != nil {
			return fmt.Errorf("chmod %s: %w", outPath, err)
		}
	}

	fmt.Printf("installed grd %s to %s\n", version, outPath)
	return nil
}

func getGrdVersion() (string, error) {
	out, err := exec.Command("grd", "--version").Output()
	if err != nil {
		combined, cerr := exec.Command("grd", "--version").CombinedOutput()
		if cerr != nil {
			return "", fmt.Errorf("grd --version failed: %w", cerr)
		}
		out = combined
	}
	s := strings.TrimSpace(string(out))
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return "", fmt.Errorf("unexpected grd --version output: %q", s)
	}
	return parts[len(parts)-1], nil
}

func goosArchToRustTarget(goos, goarch string) (string, bool) {
	switch {
	case goos == "linux" && goarch == "amd64":
		return "x86_64-unknown-linux-musl", true
	case goos == "linux" && goarch == "arm64":
		return "aarch64-unknown-linux-musl", true
	case goos == "darwin" && goarch == "arm64":
		return "aarch64-apple-darwin", true
	case goos == "windows" && goarch == "amd64":
		return "x86_64-pc-windows-gnu", true
	default:
		return "", false
	}
}

func normalizeV(v string) string {
	if !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

var getGrdCmd = &cobra.Command{
	Use:   "get-grd",
	Short: "Install or upgrade grd to the specified version",
	Args:  cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		return runGetGrd()
	},
}

func init() {
	rootCmd.AddCommand(getGrdCmd)
	getGrdCmd.Flags().StringVarP(&getGrdVer, "version", "V", "0.9.1", "Version of grd to install")
	getGrdCmd.Flags().
		BoolVarP(&getGrdLatest, "latest", "U", false, "Install or upgrade to the latest release")
	getGrdCmd.MarkFlagsMutuallyExclusive("version", "latest")
}
