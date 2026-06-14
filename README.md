# i

**Global Installer Manager** — a uniform CLI interface to multiple package managers with global install features.

Manage globally-installed CLI tools across different ecosystems from a single command.

## Features

- **Unified workflow** — add, remove, list, sync, and upgrade tools regardless of their underlying package manager
- **Multi-manager** — supports `cargo` (via `cargo-binstall`), `uv` (Python tools), and `bun` (JS/TS)
- **Declarative config** — installed packages tracked in `~/.i/packages.toml`
- **Plugin drivers** — easy to extend with new package managers

## Installation

```bash
go install github.com/lucidfrontier45/i@latest
```

### Dependencies

| Manager | Requirement                          |
| :------ | :----------------------------------- |
| `cargo` | [`cargo-binstall`](https://github.com/cargo-bins/cargo-binstall) |
| `uv`    | [`uv`](https://docs.astral.sh/uv/)   |
| `bun`   | [`bun`](https://bun.sh)                        |

## Usage

### Add a package

```bash
i add <package> --manager <manager> [--version <version>]
```

Install a package and register it in the config.

```bash
i add starship --manager cargo
i add ruff --manager uv --version 0.11.0
```

Some managers support features/extras using bracket syntax:

```bash
i add "pandas[performance,plot]" --manager uv
```

### Remove a package

```bash
i remove <package>
```

Uninstall and deregister.

### List registered packages

```bash
i list
```

### Sync all packages

```bash
i sync
```

Install all registered packages at their specified versions. Use `--force` / `-f` to reinstall even if already present:

```bash
i sync --force
```

### Upgrade packages

```bash
i upgrade       # upgrade all registered packages
i upgrade <pkg> # upgrade a specific package
```

## Supported managers

| Manager | Name     | Status | Command used              |
| :------ | :------- | :----- | :------------------------ |
| Cargo   | `cargo`  | ✅     | `cargo binstall`          |
| uv      | `uv`     | ✅     | `uv tool install`         |
| Bun     | `bun`    | ✅     | `bun i -g`                |

## How it works

Each package manager implements the `types.Driver` interface (`internal/types/package.go`). The CLI delegates install/upgrade/remove operations to the appropriate driver. Package metadata is persisted in `~/.i/packages.toml`.

```toml
[packages]
starship = { manager = "cargo", version = "1.22.1" }
ruff = { manager = "uv", version = "0.11.0" }
```

## Extending

To add a new package manager, implement the `types.Driver` interface and register it in an `init()` function inside the `manager` package. See `internal/manager/cargo.go` for a reference implementation.

## License

MIT
