package manager

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/lucidfrontier45/i/internal/types"
)

type npmDriver struct{}

func init() {
	Register(&npmDriver{})
}

func (n *npmDriver) Name() string {
	return "npm"
}

func (n *npmDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	args := []string{"install", "-g", pinVersion(string(spec.Name), spec.Version)}
	if forceOption(spec.Options) {
		args = append(args, "--force")
	}
	return runTool(ctx, "npm", "install", args...)
}

func (n *npmDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	return runTool(ctx, "npm", "upgrade", "install", "-g", string(spec.Name)+"@latest")
}

func (n *npmDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	return runTool(ctx, "npm", "uninstall", "uninstall", "-g", string(spec.Name))
}

func (n *npmDriver) InstalledVersion(ctx context.Context, spec types.PackageSpec) (string, error) {
	out, err := cmdOutput(ctx, "npm", "ls", "-g", string(spec.Name), "--json", "--depth=0")
	if err != nil {
		return "", fmt.Errorf("npm ls %s: %w", spec.Name, err)
	}
	return parseNpmLsJSON(string(out), string(spec.Name)), nil
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
