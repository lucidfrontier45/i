package manager

import (
	"context"

	"github.com/lucidfrontier45/i/internal/types"
)

type bunDriver struct{}

func init() {
	Register(&bunDriver{})
}

func (b *bunDriver) Name() string {
	return "bun"
}

func (b *bunDriver) Detect() bool {
	// TODO: check if bun is on PATH
	return false
}

func (b *bunDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	// TODO: exec bun add -g <pkg>@<version>
	_ = ctx
	_ = spec
	return nil
}

func (b *bunDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	// TODO: exec bun add -g <pkg>@latest
	_ = ctx
	_ = spec
	return nil
}

func (b *bunDriver) Remove(ctx context.Context, spec types.PackageSpec) error {
	// TODO: exec bun remove -g <pkg>
	_ = ctx
	_ = spec
	return nil
}

func (b *bunDriver) Installed(ctx context.Context) (map[string]string, error) {
	// TODO: parse bun pm ls output
	_ = ctx
	return nil, nil
}

func (b *bunDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	// TODO: parse bun pm ls <pkg> output
	_ = ctx
	_ = pkg
	return "", nil
}
