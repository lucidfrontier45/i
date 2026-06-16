package config

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/lucidfrontier45/i/internal/types"
)

type PackageEntry struct {
	Manager  types.ManagerType `toml:"manager"`
	Version  string            `toml:"version"`
	Features []string          `toml:"features,omitempty"`
	With     []string          `toml:"with,omitempty"`
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

	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	err = toml.NewEncoder(f).Encode(cfg)
	if err != nil {
		return "", err
	}

	return path, nil
}
