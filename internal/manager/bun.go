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

func (b *bunDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	out, err := exec.CommandContext(
		ctx, "bun", "info", "-g", pkg, "version",
	).CombinedOutput()
	if err != nil {
		return "", nil
	}
	return strings.TrimSpace(string(out)), nil
}
