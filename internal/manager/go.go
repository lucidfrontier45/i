package manager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lucidfrontier45/i/internal/types"
)

type goDriver struct{}

func init() {
	Register(&goDriver{})
}

func (g *goDriver) Name() string {
	return "go"
}

func (g *goDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	version := spec.Version
	if version == "" {
		version = "latest"
	}
	return runTool(ctx, "go", "install", "install", pinVersion(string(spec.Name), version))
}

func (g *goDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	return runTool(ctx, "go", "upgrade", "install", pinVersion(string(spec.Name), "latest"))
}

func (g *goDriver) Remove(_ context.Context, spec types.PackageSpec) error {
	binName := binNameFromSpec(spec)
	path, err := exec.LookPath(binName)
	if err != nil {
		return fmt.Errorf("go binary %q not found on PATH: %w", binName, err)
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove go binary: %w", err)
	}
	return nil
}

func (g *goDriver) InstalledVersion(ctx context.Context, spec types.PackageSpec) (string, error) {
	binName := binNameFromSpec(spec)
	path, err := exec.LookPath(binName)
	if err != nil {
		return "", fmt.Errorf("go binary %q not found on PATH: %w", binName, err)
	}
	out, err := cmdOutput(ctx, "go", "version", "-m", path)
	if err != nil {
		return "", fmt.Errorf("go version -m %s: %w", path, err)
	}
	return parseGoVersionM(string(out)), nil
}

func binNameFromSpec(spec types.PackageSpec) string {
	if name, ok := spec.Options["bin-name"].(string); ok && name != "" {
		return name
	}
	return filepath.Base(string(spec.Name))
}

func parseGoVersionM(output string) string {
	for _, line := range strings.Split(output, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "mod\t") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 3 {
			version := fields[2]
			if version != "" {
				return version
			}
		}
	}
	return ""
}
