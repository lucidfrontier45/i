package types

import "context"

type PackageSpec struct {
	Name    string
	Manager string
	Version string
	Options map[string]any
}

type Driver interface {
	Name() string

	Detect() bool

	Install(ctx context.Context, spec PackageSpec) error

	Upgrade(ctx context.Context, spec PackageSpec) error

	Remove(ctx context.Context, spec PackageSpec) error

	Installed(ctx context.Context) (map[string]string, error)

	InstalledVersion(ctx context.Context, pkg string) (string, error)
}
