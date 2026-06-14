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

func (u *uvDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	args := []string{"tool", "install"}
	if spec.Version != "" {
		args = append(args, spec.Name+"=="+spec.Version)
	} else {
		args = append(args, spec.Name)
	}
	if force, ok := spec.Options["force"].(bool); ok && force {
		args = append(args, "--reinstall")
	}
	out, err := exec.CommandContext(ctx, "uv", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("uv tool install: %w\n%s", err, string(out))
	}
	return nil
}

func (u *uvDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	out, err := exec.CommandContext(ctx, "uv", "tool", "install", "--upgrade", spec.Name).
		CombinedOutput()
	if err != nil {
		return fmt.Errorf("uv tool upgrade: %w\n%s", err, string(out))
	}
	return nil
}

func (u *uvDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	out, err := exec.CommandContext(ctx, "uv", "tool", "uninstall", spec.Name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("uv tool uninstall: %w\n%s", err, string(out))
	}
	return nil
}

func (u *uvDriver) Installed(ctx context.Context) (map[string]string, error) {
	out, err := exec.CommandContext(ctx, "uv", "tool", "list").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("uv tool list: %w", err)
	}
	return parseUvToolList(string(out)), nil
}

func (u *uvDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	installed, err := u.Installed(ctx)
	if err != nil {
		return "", err
	}
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
