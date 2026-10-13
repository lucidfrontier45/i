package cmd

import (
	"fmt"
	"sort"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/manager"
	"github.com/lucidfrontier45/i/internal/types"
)

// readConfig loads the config, wrapping read failures with a uniform message.
func readConfig() (*config.Config, string, error) {
	cfg, path, err := config.Read()
	if err != nil {
		return nil, "", fmt.Errorf("read config: %w", err)
	}
	return cfg, path, nil
}

// writeConfig persists cfg, wrapping write failures with a uniform message.
func writeConfig(cfg *config.Config) (string, error) {
	path, err := config.Write(cfg)
	if err != nil {
		return "", fmt.Errorf("write config: %w", err)
	}
	return path, nil
}

// lookupDriver resolves a manager type to its driver.
func lookupDriver(name types.ManagerType) (types.Driver, error) {
	drv := manager.Lookup(string(name))
	if drv == nil {
		return nil, fmt.Errorf("unknown manager %q", name)
	}
	return drv, nil
}

// requirePackage resolves key (an alias or full name) and returns the full name
// together with its entry, erroring when the package is not registered.
func requirePackage(
	cfg *config.Config,
	key string,
) (types.PackageName, config.PackageEntry, error) {
	full := cfg.ResolveName(key)
	entry, ok := cfg.Packages[full]
	if !ok {
		return "", config.PackageEntry{}, fmt.Errorf("package %q not found in config", key)
	}
	return full, entry, nil
}

// selectPackages returns the canonical names a command should act on: the
// resolved subset when keys is non-empty, otherwise every registered package in
// ascending order.
func selectPackages(
	cfg *config.Config,
	keys []string,
) ([]types.PackageName, error) {
	if len(keys) == 0 {
		return sortedNames(cfg.Packages), nil
	}
	return resolvePackageNames(cfg, keys)
}

// resolvePackageNames maps CLI keys (aliases or full names) to canonical package
// names, erroring on the first unknown key. Duplicates are collapsed and the
// result is sorted so a repeated key never acts twice and output is stable.
func resolvePackageNames(
	cfg *config.Config,
	keys []string,
) ([]types.PackageName, error) {
	names := make([]types.PackageName, 0, len(keys))
	seen := make(map[types.PackageName]bool, len(keys))
	for _, key := range keys {
		full, _, err := requirePackage(cfg, key)
		if err != nil {
			return nil, err
		}
		if seen[full] {
			continue
		}
		seen[full] = true
		names = append(names, full)
	}
	sort.Slice(names, func(i, j int) bool { return names[i] < names[j] })
	return names, nil
}
