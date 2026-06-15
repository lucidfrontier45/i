package manager

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/lucidfrontier45/i/internal/types"
)

type npmDriver struct{}

func init() {
	Register(&npmDriver{})
}

func (n *npmDriver) Name() string {
	return "npm"
}

func (n *npmDriver) Detect() bool {
	return exec.Command("npm", "--version").Run() == nil
}

func (n *npmDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	args := []string{"install", "-g"}
	if spec.Version != "" {
		args = append(args, spec.Name+"@"+spec.Version)
	} else {
		args = append(args, spec.Name)
	}
	if force, ok := spec.Options["force"].(bool); ok && force {
		args = append(args, "--force")
	}
	out, err := cmdOutput(ctx, "npm", args...)
	if err != nil {
		return fmt.Errorf("npm install: %w\n%s", err, string(out))
	}
	return nil
}

func (n *npmDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	out, err := cmdOutput(ctx, "npm", "install", "-g", spec.Name+"@latest")
	if err != nil {
		return fmt.Errorf("npm upgrade: %w\n%s", err, string(out))
	}
	return nil
}

func (n *npmDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	out, err := cmdOutput(ctx, "npm", "uninstall", "-g", spec.Name)
	if err != nil {
		return fmt.Errorf("npm uninstall: %w\n%s", err, string(out))
	}
	return nil
}

func (n *npmDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	out, err := cmdOutput(ctx, "npm", "ls", "-g", pkg, "version")
	if err != nil {
		return "", nil
	}
	return parseNpmLsVersion(string(out), pkg), nil
}

func parseNpmLsVersion(output, pkg string) string {
	prefix := pkg + "@"
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := strings.TrimPrefix(line, prefix)
		if idx := strings.IndexAny(rest, " "); idx >= 0 {
			rest = rest[:idx]
		}
		return rest
	}
	return ""
}
