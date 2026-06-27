package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/lucidfrontier45/i/internal/types"
)

type PackageEntry struct {
	Manager  types.ManagerType `toml:"manager"`
	Version  string            `toml:"version"`
	Features []string          `toml:"features,omitempty"`
	Options  map[string]any    `toml:"options,omitempty"`
}

// Config holds the TOML configuration mapping aliases to full package names
// (Index) and full package names to their package entries (Packages).
type Config struct {
	Index    map[types.PackageAlias]types.PackageName `toml:"index,omitempty"`
	Packages map[types.PackageName]PackageEntry       `toml:"packages"`
}

// ResolveName maps a user-supplied key (which may be an alias) to the full
// package name. If key is an alias in Index, the mapped full name is returned;
// otherwise key is treated as the full package name itself.
func (c *Config) ResolveName(key string) types.PackageName {
	if full, ok := c.Index[types.PackageAlias(key)]; ok && full != "" {
		return full
	}
	return types.PackageName(key)
}

func Read() (*Config, string, error) {
	path, err := Path()
	if err != nil {
		return nil, "", err
	}

	cfg := &Config{
		Index:    make(map[types.PackageAlias]types.PackageName),
		Packages: make(map[types.PackageName]PackageEntry),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, path, nil
		}
		return nil, "", err
	}

	err = toml.Unmarshal(data, cfg)
	if err != nil {
		return nil, "", err
	}

	return cfg, path, nil
}

func Write(cfg *Config) (string, error) {
	path, err := Path()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".packages.toml-*")
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()

	enc := toml.NewEncoder(tmp)
	enc.Indent = ""
	if err := enc.Encode(cfg); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := os.Chmod(tmpPath, 0o644); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return path, nil
}
