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
		args = append(args, string(spec.Name)+"@"+spec.Version)
	} else {
		args = append(args, string(spec.Name))
	}
	if force, ok := spec.Options["force"].(bool); ok && force {
		args = append(args, "--force")
	}
	out, err := cmdOutput(ctx, "bun", args...)
	if err != nil {
		return fmt.Errorf("bun install: %w\n%s", err, string(out))
	}
	return nil
}

func (b *bunDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	out, err := cmdOutput(ctx, "bun", "i", "-g", string(spec.Name)+"@latest")
	if err != nil {
		return fmt.Errorf("bun upgrade: %w\n%s", err, string(out))
	}
	return nil
}

func (b *bunDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	out, err := cmdOutput(ctx, "bun", "uninstall", "-g", string(spec.Name))
	if err != nil {
		return fmt.Errorf("bun uninstall: %w\n%s", err, string(out))
	}
	return nil
}

func (b *bunDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	out, err := cmdOutput(ctx, "bun", "pm", "ls", "-g")
	if err != nil {
		return "", fmt.Errorf("bun pm ls: %w", err)
	}
	prefix := pkg + "@"
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "├── ")
		line = strings.TrimPrefix(line, "└── ")
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimPrefix(line, prefix), nil
		}
	}
	return "", nil
}
