package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lucidfrontier45/i/internal/types"
)

type cargoDriver struct{}

func init() {
	Register(&cargoDriver{})
}

func (c *cargoDriver) Name() string {
	return "cargo"
}

func cargoHome() string {
	if home, ok := os.LookupEnv("CARGO_HOME"); ok && home != "" {
		return home
	}
	if runtime.GOOS == "windows" {
		if p := os.Getenv("USERPROFILE"); p != "" {
			return filepath.Join(p, ".cargo")
		}
	}
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".cargo")
	}
	return ".cargo"
}

func (c *cargoDriver) runInstall(ctx context.Context, name, version string, force bool) error {
	args := []string{"binstall", "-y"}
	if version != "" {
		args = append(args, name+"@"+version)
	} else {
		args = append(args, name)
	}
	if force {
		args = append(args, "--force")
	}
	out, err := cmdOutput(ctx, "cargo", args...)
	if err != nil {
		return fmt.Errorf("cargo binstall: %w\n%s", err, string(out))
	}
	return nil
}

func (c *cargoDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	force, _ := spec.Options["force"].(bool)
	return c.runInstall(ctx, string(spec.Name), spec.Version, force)
}

func (c *cargoDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	return c.runInstall(ctx, string(spec.Name), "", false)
}

func (c *cargoDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	out, err := cmdOutput(ctx, "cargo", "uninstall", string(spec.Name))
	if err != nil {
		return fmt.Errorf("cargo uninstall: %w\n%s", err, string(out))
	}
	return nil
}

func (c *cargoDriver) InstalledVersion(
	ctx context.Context,
	spec types.PackageSpec,
) (string, error) {
	data, err := os.ReadFile(filepath.Join(cargoHome(), "binstall", "crates-v1.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read binstall metadata: %w", err)
	}
	installed := parseBinstallCrates(data)
	return installed[string(spec.Name)], nil
}

// binstallCrate is a single record from cargo-binstall's crates metadata.
type binstallCrate struct {
	Name    string `json:"name"`
	Version string `json:"current_version"`
}

func parseBinstallCrates(data []byte) map[string]string {
	if result := decodeCratesStream(data); len(result) > 0 {
		return result
	}
	// Fallback: try extracting from the raw text if JSON decoder got nothing
	return parseCratesFallback(data)
}

// decodeCratesStream decodes a stream of concatenated JSON crate records.
func decodeCratesStream(data []byte) map[string]string {
	result := make(map[string]string)
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		var entry binstallCrate
		if err := dec.Decode(&entry); err != nil {
			break
		}
		recordCrate(result, entry)
	}
	return result
}

// parseCratesFallback recovers crate records from a raw text blob by splitting
// on the `}{` boundary, for inputs the JSON stream decoder cannot handle.
func parseCratesFallback(data []byte) map[string]string {
	result := make(map[string]string)
	for _, block := range strings.Split(string(data), `}{`) {
		block = strings.TrimSpace(block)
		if block == "" {
			continue
		}
		if !strings.HasPrefix(block, "{") {
			block = "{" + block
		}
		if !strings.HasSuffix(block, "}") {
			block = block + "}"
		}
		var entry binstallCrate
		if err := json.Unmarshal([]byte(block), &entry); err != nil {
			continue
		}
		recordCrate(result, entry)
	}
	return result
}

// recordCrate stores a non-empty crate record, keeping the first version seen
// for a given name.
func recordCrate(result map[string]string, entry binstallCrate) {
	if entry.Name == "" || entry.Version == "" {
		return
	}
	if _, exists := result[entry.Name]; !exists {
		result[entry.Name] = entry.Version
	}
}
