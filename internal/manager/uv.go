package manager

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/lucidfrontier45/i/internal/types"
)

type uvDriver struct{}

func init() {
	Register(&uvDriver{})
}

func (u *uvDriver) Name() string {
	return "uv"
}

func (u *uvDriver) Detect() bool {
	return exec.Command("uv", "--version").Run() == nil
}

func installPkgName(name string, features []string) string {
	if len(features) == 0 {
		return name
	}
	return name + "[" + strings.Join(features, ",") + "]"
}

func (u *uvDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	pkg := installPkgName(string(spec.Name), spec.Features)
	args := []string{"tool", "install"}
	if spec.Version != "" {
		args = append(args, pkg+"=="+spec.Version)
	} else {
		args = append(args, pkg)
	}
	for _, w := range spec.With {
		args = append(args, "--with", w)
	}
	if force, ok := spec.Options["force"].(bool); ok && force {
		args = append(args, "--reinstall")
	}
	out, err := cmdOutput(ctx, "uv", args...)
	if err != nil {
		return fmt.Errorf("uv tool install: %w\n%s", err, string(out))
	}
	return nil
}

func (u *uvDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	pkg := installPkgName(string(spec.Name), spec.Features)
	args := []string{"tool", "install", "--upgrade", pkg}
	for _, w := range spec.With {
		args = append(args, "--with", w)
	}
	out, err := cmdOutput(ctx, "uv", args...)
	if err != nil {
		return fmt.Errorf("uv tool upgrade: %w\n%s", err, string(out))
	}
	return nil
}

func (u *uvDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	pkg := installPkgName(string(spec.Name), spec.Features)
	out, err := cmdOutput(ctx, "uv", "tool", "uninstall", pkg)
	if err != nil {
		return fmt.Errorf("uv tool uninstall: %w\n%s", err, string(out))
	}
	return nil
}

func (u *uvDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	out, err := cmdOutput(ctx, "uv", "tool", "list")
	if err != nil {
		return "", fmt.Errorf("uv tool list: %w", err)
	}
	installed := parseUvToolList(string(out))
	return installed[pkg], nil
}

func parseUvToolList(output string) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "- ")
		parts := strings.SplitN(line, " v", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		version := strings.TrimSpace(parts[1])
		if name != "" && version != "" {
			result[name] = version
		}
	}
	return result
}
