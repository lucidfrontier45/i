# Architecture

## Overview

`i` is a unified CLI for managing globally-installed CLI tools across six package ecosystems. It wraps each ecosystem's native tool (`cargo-binstall`, `uv`, `bun`, `npm`, `grd`, `go install`) behind a common `Driver` interface and persists a declarative manifest at `~/.i/packages.toml`.

Repository: `github.com/lucidfrontier45/i`  
Language: Go 1.25.6  
License: MIT

---

## Layered Structure

```
main.go
 └─ cmd/                  CLI entry points (cobra commands)
     ├─ root.go           persistent flags, startup cleanup
     ├─ add.go            register + install a package
     ├─ remove.go         uninstall + deregister
     ├─ list.go           table of registered packages
     ├─ sync.go           install all registered packages
     ├─ upgrade.go        upgrade one or all packages
     ├─ self_upgrade.go   replace the running i binary
     └─ version.go        build info
 ├─ internal/
     ├─ config/           TOML persistence layer
     │   ├── toml.go      Config (Index + Packages), Read/Write
     │   └── path.go      resolve ~/.i/packages.toml
     ├─ types/            core domain types
     │   └── package.go   Driver interface, PackageSpec, ManagerType
     ├─ manager/          driver registry + per-ecosystem implementations
     │   ├── registry.go  Register/Lookup/All/Detect
     │   ├── verbose.go   exec helper with optional stderr printing
     │   └── <driver>.go  one file per ecosystem (cargo, bun, npm, go, uv, grd)
     └── selfupdate/      binary self-replacement
         └── update.go    GitHub releases API, checksum verify, Windows-safe replace
```

---

## Core Domain Model

| Type                           | Purpose                                                                       |
| ------------------------------ | ----------------------------------------------------------------------------- |
| `PackageName` / `PackageAlias` | Typed strings distinguishing the real package name from the user's shorthand  |
| `ManagerType`                  | Enum constant identifying the ecosystem                                       |
| `PackageSpec`                  | Full install descriptor (name, manager, version, features, free-form options) |

### `Driver` Interface — The Extension Point

```go
type Driver interface {
    Name() string
    Detect() bool
    Install(ctx, spec) error
    Upgrade(ctx, spec) error
    Remove(ctx, spec) error
    InstalledVersion(ctx, pkg) (string, error)
}
```

Every driver self-registers in its `init()` via `manager.Register`. Adding a new ecosystem requires one new file implementing these five methods.

---

## Config Persistence

**File:** `~/.i/packages.toml`

Two-section structure:

- **`Index`** — optional alias indirection (alias → full package name)
- **`Packages`** — canonical registry keyed by full name, each entry holds manager, version, features, and driver-specific options

`config.ResolveName(key)` checks aliases first; if no match, the key is used as the package name directly.

---

## CLI Data Flow

```
User args
    │
    ▼
cmd/ handler
    ├─ config.Read()          load ~/.i/packages.toml
    ├─ cfg.ResolveName(key)   alias → canonical name
    ├─ manager.Lookup(name)   resolve ManagerType string → Driver
    ├─ driver.Install/Upgrade/Remove(ctx, spec)
    └─ config.Write(cfg)      persist changes
```

After install/upgrade, `InstalledVersion` is queried and the config's `Version` field is updated to reflect the actual installed version.

---

## Driver Registry

A package-level slice populated by `init()`. Provides:

- `Register(d)` — append driver
- `Lookup(name)` — linear scan by `d.Name()`
- `Detect()` — subset whose `Detect()` returns true (tools available on PATH)
- `All()` — all registered drivers

Verbose mode is a package-level bool set by `root.go`'s `PersistentPreRun`; `cmdOutput` prints the command to stderr when enabled.

---

## Self-Update

The `self-upgrade` command downloads the latest GitHub release, verifies the SHA256 checksum against `checksums.txt`, and replaces the running binary. On Windows, the running executable is renamed to `<exe>.to_remove` first (it cannot be overwritten while running); the leftover is cleaned on the next normal invocation.

---

## Key Design Decisions

1. **No lock file** — config is read + written without atomic file locking. Concurrent invocations on the same config risk data loss.

2. **Global mutable registry** — drivers register themselves in `init()`. Simple but requires explicit state reset for testing.

3. **Config as source of truth** — the TOML file is both registry and desired-state manifest. `sync` is a declarative install-all.

4. **No tests** — the project has no `*_test.go` files across any package.

5. **Dependencies** — only three external modules: `BurntSushi/toml` (config), `spf13/cobra` (CLI), `golang.org/x/mod` (semver comparison in self-update).
