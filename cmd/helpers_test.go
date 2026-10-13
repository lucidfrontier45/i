package cmd

import (
	"slices"
	"testing"

	"github.com/lucidfrontier45/i/internal/config"
	"github.com/lucidfrontier45/i/internal/types"
)

// testConfig builds a config with two aliases pointing at ripgrep (exercising
// many-to-one aliasing) plus a package addressed only by its full name.
func testConfig() *config.Config {
	return &config.Config{
		Index: map[types.PackageAlias]types.PackageName{
			"rg":   "ripgrep",
			"grep": "ripgrep",
			"fd":   "fd-find",
		},
		Packages: map[types.PackageName]config.PackageEntry{
			"ripgrep": {Manager: types.ManagerCargo},
			"fd-find": {Manager: types.ManagerCargo},
			"jq":      {Manager: types.ManagerNpm},
		},
	}
}

func TestResolvePackageNames(t *testing.T) {
	tests := []struct {
		name    string
		keys    []string
		want    []types.PackageName
		wantErr bool
	}{
		{
			name: "full name",
			keys: []string{"jq"},
			want: []types.PackageName{"jq"},
		},
		{
			name: "alias resolves to full name",
			keys: []string{"rg"},
			want: []types.PackageName{"ripgrep"},
		},
		{
			name: "alias and full name collapse to one",
			keys: []string{"rg", "ripgrep"},
			want: []types.PackageName{"ripgrep"},
		},
		{
			name: "distinct aliases pointing at the same package collapse",
			keys: []string{"rg", "grep"},
			want: []types.PackageName{"ripgrep"},
		},
		{
			name: "result is sorted",
			keys: []string{"jq", "rg"},
			want: []types.PackageName{"jq", "ripgrep"},
		},
		{
			name:    "unknown key errors",
			keys:    []string{"nope"},
			wantErr: true,
		},
		{
			name:    "unknown key among known keys errors",
			keys:    []string{"jq", "nope"},
			wantErr: true,
		},
	}

	cfg := testConfig()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolvePackageNames(cfg, tt.keys)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolvePackageNames(%v) = %v, want error", tt.keys, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolvePackageNames(%v): unexpected error: %v", tt.keys, err)
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("resolvePackageNames(%v) = %v, want %v", tt.keys, got, tt.want)
			}
		})
	}
}

func TestSelectPackages(t *testing.T) {
	cfg := testConfig()

	t.Run("no keys selects every package sorted", func(t *testing.T) {
		got, err := selectPackages(cfg, nil)
		if err != nil {
			t.Fatalf("selectPackages(nil): %v", err)
		}
		want := []types.PackageName{"fd-find", "jq", "ripgrep"}
		if !slices.Equal(got, want) {
			t.Fatalf("selectPackages(nil) = %v, want %v", got, want)
		}
	})

	t.Run("keys select the resolved subset", func(t *testing.T) {
		got, err := selectPackages(cfg, []string{"rg", "jq"})
		if err != nil {
			t.Fatalf("selectPackages([rg jq]): %v", err)
		}
		want := []types.PackageName{"jq", "ripgrep"}
		if !slices.Equal(got, want) {
			t.Fatalf("selectPackages([rg jq]) = %v, want %v", got, want)
		}
	})
}
