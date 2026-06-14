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

// grdListLineFormat is the per-line shape of `grd --list-installed` stdout:
//
//	"<repo> (tag: <tag>, asset: <asset>)"
const grdListLinePrefix = " (tag: "

func (g *grdDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	args := []string{spec.Name}
	if spec.Version != "" {
		args = append(args, "--tag", spec.Version)
	}
	g.appendCommonFlags(&args, spec.Options)
	if force, ok := spec.Options["force"].(bool); ok && force {
		args = append(args, "--force")
	}
	out, err := exec.CommandContext(ctx, "grd", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("grd install: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (g *grdDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	// -y is required: grd's upgrade-prompt branch refuses to run when stdin
	// is not a TTY, which is always the case under `i`.
	args := []string{spec.Name, "-y"}
	g.appendCommonFlags(&args, spec.Options)
	if force, ok := spec.Options["force"].(bool); ok && force {
		args = append(args, "--force")
	}
	out, err := exec.CommandContext(ctx, "grd", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("grd upgrade: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (g *grdDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	args := []string{spec.Name, "--remove"}
	if dst, ok := spec.Options["destination"].(string); ok && dst != "" {
		args = append(args, "--destination", dst)
	}
	if name, ok := spec.Options["bin-name"].(string); ok && name != "" {
		args = append(args, "--bin-name", name)
	}
	out, err := exec.CommandContext(ctx, "grd", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("grd remove: %w\n%s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (g *grdDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	out, err := exec.CommandContext(ctx, "grd", "--list-installed").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("grd --list-installed: %w", err)
	}
	return parseGrdListInstalled(string(out), pkg), nil
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

// parseGrdListInstalled scans grd --list-installed stdout for the entry
// matching pkg and returns its tag. Returns "" when the package is absent
// or the output is empty.
func parseGrdListInstalled(output, pkg string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line == "No installed packages found." {
			continue
		}
		if !strings.HasPrefix(line, pkg+grdListLinePrefix) {
			continue
		}
		rest := line[len(pkg)+len(grdListLinePrefix):]
		end := strings.Index(rest, ",")
		if end == -1 {
			return ""
		}
		return strings.TrimSpace(rest[:end])
	}
	return ""
}
