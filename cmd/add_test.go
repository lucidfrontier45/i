package cmd

import (
	"testing"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/types"
)

func TestDeriveAlias(t *testing.T) {
	tests := []struct {
		name string
		mgr  types.ManagerType
		pkg  types.PackageName
		want types.PackageAlias
	}{
		{
			name: "grd owner/repo derives repo name",
			mgr:  types.ManagerGrd,
			pkg:  "owner/repo",
			want: "repo",
		},
		{
			name: "grd nested path derives last segment",
			mgr:  types.ManagerGrd,
			pkg:  "a/b/c",
			want: "c",
		},
		{
			name: "npm scoped package derives unqualified name",
			mgr:  types.ManagerNpm,
			pkg:  "@user/package",
			want: "package",
		},
		{
			name: "bun unscoped name has no slash",
			mgr:  types.ManagerBun,
			pkg:  "typescript",
			want: "",
		},
		{
			name: "trailing slash derives nothing",
			mgr:  types.ManagerGrd,
			pkg:  "owner/",
			want: "",
		},
		{
			name: "go manager has no shorthand rule",
			mgr:  types.ManagerGo,
			pkg:  "golang.org/x/tools/gopls",
			want: "",
		},
		{
			name: "uv manager has no shorthand rule",
			mgr:  types.ManagerUv,
			pkg:  "pandas[performance]",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := deriveAlias(tt.mgr, tt.pkg)
			if got != tt.want {
				t.Fatalf("deriveAlias(%s, %s) = %q, want %q", tt.mgr, tt.pkg, got, tt.want)
			}
			wantOK := tt.want != ""
			if ok != wantOK {
				t.Fatalf("deriveAlias(%s, %s) ok = %v, want %v", tt.mgr, tt.pkg, ok, wantOK)
			}
		})
	}
}

// autoAliasConfig is a config where the alias "grep" points at the registered
// package "ripgrep".
func autoAliasConfig() *config.Config {
	return &config.Config{
		Index: map[types.PackageAlias]types.PackageName{
			"grep": "ripgrep",
		},
		Packages: map[types.PackageName]config.PackageEntry{
			"ripgrep": {Manager: types.ManagerCargo},
		},
	}
}

func TestApplyAutoAlias(t *testing.T) {
	cfg := autoAliasConfig()

	tests := []struct {
		name  string
		alias string
		mgr   types.ManagerType
		pkg   types.PackageName
		want  types.PackageAlias
	}{
		{
			name:  "explicit alias wins",
			alias: "mine",
			mgr:   types.ManagerGrd,
			pkg:   "owner/repo",
			want:  "mine",
		},
		{
			name: "grd derives repo name",
			mgr:  types.ManagerGrd,
			pkg:  "owner/repo",
			want: "repo",
		},
		{
			name: "npm scoped derives unqualified name",
			mgr:  types.ManagerNpm,
			pkg:  "@scope/pkg",
			want: "pkg",
		},
		{
			name: "bun unscoped name stays alias-free",
			mgr:  types.ManagerBun,
			pkg:  "typescript",
			want: "",
		},
		{
			name: "non-eligible manager is untouched",
			mgr:  types.ManagerGo,
			pkg:  "golang.org/x/tools/gopls",
			want: "",
		},
		{
			name: "derived name shadowing a package is skipped",
			mgr:  types.ManagerGrd,
			pkg:  "owner/ripgrep",
			want: "",
		},
		{
			name: "derived alias bound elsewhere is skipped",
			mgr:  types.ManagerGrd,
			pkg:  "owner/grep",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			up := addUpdate{
				pkg:       tt.pkg,
				mgr:       tt.mgr,
				aliasFlag: types.PackageAlias(tt.alias),
			}
			up.applyAutoAlias(cfg)
			if up.aliasFlag != tt.want {
				t.Fatalf(
					"applyAutoAlias(%s, %s, alias %q) aliasFlag = %q, want %q",
					tt.mgr, tt.pkg, tt.alias, up.aliasFlag, tt.want,
				)
			}
		})
	}
}

func TestAliasConflict(t *testing.T) {
	cfg := autoAliasConfig()

	tests := []struct {
		name    string
		alias   types.PackageAlias
		pkg     types.PackageName
		wantErr bool
	}{
		{
			name:  "free alias is available",
			alias: "repo",
			pkg:   "owner/repo",
		},
		{
			name:  "alias bound to the same package is not a conflict",
			alias: "grep",
			pkg:   "ripgrep",
		},
		{
			name:    "alias bound to another package conflicts",
			alias:   "grep",
			pkg:     "owner/grep",
			wantErr: true,
		},
		{
			name:    "alias shadowing a package name conflicts",
			alias:   "ripgrep",
			pkg:     "owner/ripgrep",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := aliasConflict(cfg, tt.alias, tt.pkg)
			if tt.wantErr && err == nil {
				t.Fatalf("aliasConflict(%s, %s) = nil, want error", tt.alias, tt.pkg)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("aliasConflict(%s, %s): unexpected error: %v", tt.alias, tt.pkg, err)
			}
		})
	}
}
