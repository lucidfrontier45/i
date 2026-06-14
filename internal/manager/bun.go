package manager

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/lucidfrontier45/i/internal/types"
)

type bunDriver struct{}

func init() {
	Register(&bunDriver{})
}

func (b *bunDriver) Name() string {
	return "bun"
}

func (b *bunDriver) Detect() bool {
	return exec.Command("bun", "--version").Run() == nil
}

func (b *bunDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	args := []string{"i", "-g"}
	if spec.Version != "" {
		args = append(args, spec.Name+"@"+spec.Version)
	} else {
		args = append(args, spec.Name)
	}
	if force, ok := spec.Options["force"].(bool); ok && force {
		args = append(args, "--force")
	}
	out, err := exec.CommandContext(ctx, "bun", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("bun install: %w\n%s", err, string(out))
	}
	return nil
}

func (b *bunDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	out, err := exec.CommandContext(
		ctx, "bun", "i", "-g", spec.Name+"@latest",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("bun upgrade: %w\n%s", err, string(out))
	}
	return nil
}

func (b *bunDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	out, err := exec.CommandContext(
		ctx, "bun", "uninstall", "-g", spec.Name,
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("bun uninstall: %w\n%s", err, string(out))
	}
	return nil
}

func (b *bunDriver) Installed(ctx context.Context) (map[string]string, error) {
	out, err := exec.CommandContext(
		ctx, "bun", "pm", "-g", "list",
	).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("bun pm list: %w\n%s", err, string(out))
	}
	return parseBunGlobalList(string(out)), nil
}

func (b *bunDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	out, err := exec.CommandContext(
		ctx, "bun", "info", "-g", pkg, "version",
	).CombinedOutput()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}

// parseBunGlobalList parses "bun pm -g list" output.
//
// Input is a tree listing:
//
//	D:\toolchains\bun\install\global node_modules (496)
//	├── pkg@1.0.0
//	├── @scope/name@2.0.0
//	└── last@3.0.0
func parseBunGlobalList(output string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var pkgLine string
		switch {
		case strings.HasPrefix(line, "├── "):
			pkgLine = line[4:]
		case strings.HasPrefix(line, "└── "):
			pkgLine = line[4:]
		default:
			continue
		}
		// Split on last '@' to handle scoped packages like @scope/name@1.0.0.
		idx := strings.LastIndexByte(pkgLine, '@')
		if idx == -1 {
			continue
		}
		name := pkgLine[:idx]
		version := pkgLine[idx+1:]
		if name != "" && version != "" {
			result[name] = version
		}
	}
	return result
}
