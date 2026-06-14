package manager

import (
	"context"

	"github.com/lucidfrontier45/i/internal/types"
)

type uvDriver struct{}

func init() {
	Register(&uvDriver{})
}

func (u *uvDriver) Name() string {
	return "uv"
}

func (u *uvDriver) Detect() bool {
	// TODO: check if uv is on PATH
	return false
}

func (u *uvDriver) Install(ctx context.Context, spec types.PackageSpec) error {
	// TODO: exec uv tool install <pkg>==<version>
	_ = ctx
	_ = spec
	return nil
}

func (u *uvDriver) Upgrade(ctx context.Context, spec types.PackageSpec) error {
	// TODO: exec uv tool install --upgrade <pkg>
	_ = ctx
	_ = spec
	return nil
}

func (u *uvDriver) Delete(ctx context.Context, spec types.PackageSpec) error {
	// TODO: exec uv tool uninstall <pkg>
	_ = ctx
	_ = spec
	return nil
}

func (u *uvDriver) Installed(ctx context.Context) (map[string]string, error) {
	// TODO: parse uv tool list output
	_ = ctx
	return nil, nil
}

func (u *uvDriver) InstalledVersion(ctx context.Context, pkg string) (string, error) {
	// TODO: parse uv tool list output for specific pkg
	_ = ctx
	_ = pkg
	return "", nil
}
