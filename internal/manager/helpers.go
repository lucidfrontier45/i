package manager

import (
	"context"
	"fmt"
)

// runTool executes tool and wraps a failure with the given action label and the
// captured output, giving every driver a consistent error shape.
func runTool(ctx context.Context, tool, action string, args ...string) error {
	out, err := cmdOutput(ctx, tool, args...)
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", tool, action, err, string(out))
	}
	return nil
}

// forceOption reports whether the driver-specific "force" option is set.
func forceOption(opts map[string]any) bool {
	force, _ := opts["force"].(bool)
	return force
}

// pinVersion appends "@version" to name, or returns name unchanged when version
// is empty.
func pinVersion(name, version string) string {
	if version == "" {
		return name
	}
	return name + "@" + version
}
