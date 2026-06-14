package manager

import (
	"context"

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
	// TODO: check if cargo-binstall is on PATH
	return false
}

func (c *cargoDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	// TODO: exec cargo binstall <pkg>@<version>
	_ = ctx
	_ = spec
	return nil
}

func (c *cargoDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	// TODO: exec cargo binstall <pkg>@latest
	_ = ctx
	_ = spec
	return nil
}

func (c *cargoDriver) Delete(ctx context.Context, spec types.PackageSpec) error {
	// TODO: exec cargo uninstall <pkg>
	_ = ctx
	_ = spec
	return nil
}

func (c *cargoDriver) Installed(ctx context.Context) (map[string]string, error) {
	// TODO: parse cargo install --list output
	_ = ctx
	return nil, nil
}

func (c *cargoDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	// TODO: parse cargo install --list for specific pkg
	_ = ctx
	_ = pkg
	return "", nil
}
