package manager

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/lucidfrontier45/i/internal/types"
)

type grdDriver struct{}

func init() {
	Register(&grdDriver{})
}

func (g *grdDriver) Name() string {
	return "grd"
}

func (g *grdDriver) Detect() bool {
	return exec.Command("grd", "--version").Run() == nil
}

func (g *grdDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	args := []string{string(spec.Name), "-y"}
	if spec.Version != "" {
		args = append(args, "--tag", spec.Version)
	}
	g.appendCommonFlags(&args, spec.Options)
	if force, ok := spec.Options["force"].(bool); ok && force {
		args = append(args, "--force")
	}
	out, err := cmdOutput(ctx, "grd", args...)
	if err != nil {
		return fmt.Errorf("grd install: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (g *grdDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	// -y is required: grd's prompt branches refuse to run when stdin
	// is not a TTY, which is always the case under `i`.
	args := []string{string(spec.Name), "-y"}
	g.appendCommonFlags(&args, spec.Options)
	if force, ok := spec.Options["force"].(bool); ok && force {
		args = append(args, "--force")
	}
	out, err := cmdOutput(ctx, "grd", args...)
	if err != nil {
		return fmt.Errorf("grd upgrade: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (g *grdDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	args := []string{"remove", string(spec.Name)}
	if dst, ok := spec.Options["destination"].(string); ok && dst != "" {
		args = append(args, "--destination", dst)
	}
	if name, ok := spec.Options["bin-name"].(string); ok && name != "" {
		args = append(args, "--bin-name", name)
	}
	out, err := cmdOutput(ctx, "grd", args...)
	if err != nil {
		return fmt.Errorf("grd remove: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (g *grdDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	out, err := cmdOutput(ctx, "grd", "info", pkg)
	if err != nil {
		return "", fmt.Errorf("grd info %s: %w", pkg, err)
	}
	return parseGrdInfo(string(out)), nil
}

// appendCommonFlags adds destination, bin-name, and exclude from options
// when they are present and non-empty.
func (g *grdDriver) appendCommonFlags(args *[]string, opts map[string]any) {
	if dst, ok := opts["destination"].(string); ok && dst != "" {
		*args = append(*args, "--destination", dst)
	}
	if name, ok := opts["bin-name"].(string); ok && name != "" {
		*args = append(*args, "--bin-name", name)
	}
	if excl, ok := opts["exclude"].(string); ok && excl != "" {
		*args = append(*args, "--exclude", excl)
	}
}

// parseGrdInfo extracts the tag from `grd info <pkg>` output.
// Output format: repo=owner/repo;tag=v1.0.0;asset=...;destination=...;binary=...;binary_exists=true
func parseGrdInfo(output string) string {
	for _, field := range strings.Split(strings.TrimSpace(output), ";") {
		kv := strings.SplitN(field, "=", 2)
		if len(kv) == 2 && kv[0] == "tag" {
			return kv[1]
		}
	}
	return ""
}
