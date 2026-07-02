package types

import "context"

type (
	PackageName  string
	PackageAlias string
)

type ManagerType string

const (
	ManagerBun   ManagerType = "bun"
	ManagerNpm   ManagerType = "npm"
	ManagerGo    ManagerType = "go"
	ManagerCargo ManagerType = "cargo"
	ManagerUv    ManagerType = "uv"
	ManagerGrd   ManagerType = "grd"
)

type PackageSpec struct {
	Name     PackageName
	Manager  ManagerType
	Version  string
	Features []string
	Options  map[string]any
}

type Driver interface {
	Name() string

	Install(ctx context.Context, spec PackageSpec) error

	Upgrade(ctx context.Context, spec PackageSpec) error

	Remove(ctx context.Context, spec PackageSpec) error

	InstalledVersion(ctx context.Context, spec PackageSpec) (string, error)
}
