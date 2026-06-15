package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type PackageEntry struct {
	Manager  string         `toml:"manager"`
	Version  string         `toml:"version"`
	Package  string         `toml:"package"`
	Features []string       `toml:"features,omitempty"`
	Options  map[string]any `toml:"options,omitempty"`
}

// UpstreamName returns the upstream package name to pass to a driver.
// Falls back to fallback when Package is empty (e.g. legacy configs written
// before the --alias feature was added).
func (e PackageEntry) UpstreamName(fallback string) string {
	if e.Package != "" {
		return e.Package
	}
	return fallback
}

type Config struct {
	Packages map[string]PackageEntry `toml:"packages"`
}

func Read() (*Config, string, error) {
	path, err := Path()
	if err != nil {
		return nil, "", err
	}

	cfg := &Config{Packages: make(map[string]PackageEntry)}

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
