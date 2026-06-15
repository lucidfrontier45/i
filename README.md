# i

**Global Installer Manager** — a uniform CLI interface to multiple package managers with global install features.

Manage globally-installed CLI tools across different ecosystems from a single command.

## Features

- **Unified workflow** — add, remove, list, sync, and upgrade tools regardless of their underlying package manager
- **Multi-manager** — supports `cargo` (via `cargo-binstall`), `uv` (Python tools), `bun` (JS/TS), `grd` (GitHub release binaries), and `go` (Go toolchain)
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
| `grd`   | [`grd`](https://github.com/lucidfrontier45/grd) |
| `go`    | [Go toolchain](https://go.dev/dl/)   |

## Usage

### Add a package

```bash
i add <package> --manager <manager> [--version <version>]
```

Install a package and register it in the config.

```bash
i add starship --manager cargo
i add ruff --manager uv --version 0.11.0
i add lucidfrontier45/grd --manager grd
i add golang.org/x/tools/gopls --manager go
```

Some managers support features/extras using bracket syntax:

```bash
i add "pandas[performance,plot]" --manager uv
```

#### grd-specific options

Pass these flags to `i add` when using `--manager grd`:

| Flag               | Description                                        |
| :----------------- | :------------------------------------------------- |
| `--destination`    | Destination directory for the binary               |
| `--bin-name`       | Override the binary name after download            |
| `--exclude`        | Comma-separated asset-name substrings to exclude   |

```bash
i add neovim/neovim --manager grd --destination /usr/local/bin --exclude windows
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

### Global flags

| Flag              | Description                                        |
| :---------------- | :------------------------------------------------- |
| `--verbose`, `-v` | Print every underlying command before it runs      |

The verbose flag is available on every command:

```bash
$ i -v add starship --manager cargo
+ cargo binstall starship --no-confirm
$ i -v sync
+ cargo binstall starship --no-confirm
+ uv tool install ruff
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

### Self-upgrade

```bash
i self-upgrade
```

Download and install the latest `i` release, replacing the running binary. The
download is verified against the published `checksums.txt`. On Windows the
running executable is renamed out of the way first (it cannot be overwritten
while running); the leftover is cleaned up on the next run.

### Print version

```bash
i version
```

## Supported managers

| Manager | Name     | Status | Command used              |
| :------ | :------- | :----- | :------------------------ |
| Cargo   | `cargo`  | ✅     | `cargo binstall`          |
| uv      | `uv`     | ✅     | `uv tool install`         |
| Bun     | `bun`    | ✅     | `bun i -g`                |
| grd     | `grd`    | ✅     | `grd`                     |
| Go      | `go`     | ✅     | `go install pkg@version`  |

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
