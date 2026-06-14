package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

func (c *cargoDriver) Detect() bool {
	return exec.Command("cargo", "binstall", "--help").Run() == nil
}

func cargoHome() string {
	if home, ok := os.LookupEnv("CARGO_HOME"); ok {
		return home
	}
	return filepath.Join(os.Getenv("USERPROFILE"), ".cargo")
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
	out, err := exec.CommandContext(ctx, "cargo", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cargo binstall: %w\n%s", err, string(out))
	}
	return nil
}

func (c *cargoDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	force, _ := spec.Options["force"].(bool)

	if err := c.runInstall(ctx, spec.Name, spec.Version, force); err != nil {
		return err
	}

	if _, err := exec.LookPath(spec.Name); err != nil {
		return c.runInstall(ctx, spec.Name, spec.Version, true)
	}

	return nil
}

func (c *cargoDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	if err := c.runInstall(ctx, spec.Name, "", false); err != nil {
		return err
	}

	if _, err := exec.LookPath(spec.Name); err != nil {
		return c.runInstall(ctx, spec.Name, "", true)
	}

	return nil
}

func (c *cargoDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	out, err := exec.CommandContext(ctx, "cargo", "uninstall", spec.Name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("cargo uninstall: %w\n%s", err, string(out))
	}
	return nil
}

func (c *cargoDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	data, err := os.ReadFile(filepath.Join(cargoHome(), "binstall", "crates-v1.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read binstall metadata: %w", err)
	}
	installed := parseBinstallCrates(data)
	return installed[pkg], nil
}

func parseBinstallCrates(data []byte) map[string]string {
	result := make(map[string]string)
	dec := json.NewDecoder(bytes.NewReader(data))
	for {
		var entry struct {
			Name    string `json:"name"`
			Version string `json:"current_version"`
		}
		if err := dec.Decode(&entry); err != nil {
			break
		}
		if entry.Name != "" && entry.Version != "" {
			if _, exists := result[entry.Name]; !exists {
				result[entry.Name] = entry.Version
			}
		}
	}
	// Fallback: try extracting from the raw text if JSON decoder got nothing
	if len(result) == 0 {
		raw := string(data)
		for _, block := range strings.Split(raw, `}{`) {
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
			var entry struct {
				Name    string `json:"name"`
				Version string `json:"current_version"`
			}
			if err := json.Unmarshal([]byte(block), &entry); err != nil {
				continue
			}
			if entry.Name != "" && entry.Version != "" {
				if _, exists := result[entry.Name]; !exists {
					result[entry.Name] = entry.Version
				}
			}
		}
	}
	return result
}
