package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

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
		args = append(args, string(spec.Name)+"@"+spec.Version)
	} else {
		args = append(args, string(spec.Name))
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
	out, err := cmdOutput(ctx, "npm", "install", "-g", string(spec.Name)+"@latest")
	if err != nil {
		return fmt.Errorf("npm upgrade: %w\n%s", err, string(out))
	}
	return nil
}

func (n *npmDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	out, err := cmdOutput(ctx, "npm", "uninstall", "-g", string(spec.Name))
	if err != nil {
		return fmt.Errorf("npm uninstall: %w\n%s", err, string(out))
	}
	return nil
}

func (n *npmDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	out, err := cmdOutput(ctx, "npm", "ls", "-g", pkg, "--json", "--depth=0")
	if err != nil {
		return "", fmt.Errorf("npm ls %s: %w", pkg, err)
	}
	return parseNpmLsJSON(string(out), pkg), nil
}

func parseNpmLsJSON(output, pkg string) string {
	var doc struct {
		Dependencies map[string]struct {
			Version string `json:"version"`
		} `json:"dependencies"`
	}
	if err := json.Unmarshal([]byte(output), &doc); err != nil {
		return ""
	}
	if d, ok := doc.Dependencies[pkg]; ok {
		return d.Version
	}
	return ""
}
