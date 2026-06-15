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

func (g *goDriver) Detect() bool {
	return exec.Command("go", "version").Run() == nil
}

func (g *goDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	version := spec.Version
	if version == "" {
		version = "latest"
	}
	pkg := spec.Name + "@" + version
	out, err := cmdOutput(ctx, "go", "install", pkg)
	if err != nil {
		return fmt.Errorf("go install: %w\n%s", err, string(out))
	}
	return nil
}

func (g *goDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	pkg := spec.Name + "@latest"
	out, err := cmdOutput(ctx, "go", "install", pkg)
	if err != nil {
		return fmt.Errorf("go upgrade: %w\n%s", err, string(out))
	}
	return nil
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

func (g *goDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	binName := filepath.Base(pkg)
	path, err := exec.LookPath(binName)
	if err != nil {
		return "", nil
	}
	out, err := cmdOutput(ctx, "go", "version", "-m", path)
	if err != nil {
		return "", nil
	}
	return parseGoVersionM(string(out)), nil
}

func binNameFromSpec(spec types.PackageSpec) string {
	if name, ok := spec.Options["bin-name"].(string); ok && name != "" {
		return name
	}
	return filepath.Base(spec.Name)
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
