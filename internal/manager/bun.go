package manager

import (
	"context"
	"fmt"
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

func (b *bunDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	args := []string{"i", "-g", pinVersion(string(spec.Name), spec.Version)}
	if forceOption(spec.Options) {
		args = append(args, "--force")
	}
	return runTool(ctx, "bun", "install", args...)
}

func (b *bunDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	return runTool(ctx, "bun", "upgrade", "i", "-g", string(spec.Name)+"@latest")
}

func (b *bunDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	return runTool(ctx, "bun", "uninstall", "uninstall", "-g", string(spec.Name))
}

func (b *bunDriver) InstalledVersion(ctx context.Context, spec types.PackageSpec) (string, error) {
	out, err := cmdOutput(ctx, "bun", "pm", "ls", "-g")
	if err != nil {
		return "", fmt.Errorf("bun pm ls: %w", err)
	}
	prefix := string(spec.Name) + "@"
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
