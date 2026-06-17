<img src="logo.png" alt="logo" width="500"/>

---

A unified CLI interface to multiple package managers with global install features.

Manage globally-installed CLI tools across different ecosystems from a single command.

## Features

- **Unified workflow** — add, remove, list, sync, and upgrade tools regardless of their underlying package manager
- **Multi-manager** — supports `cargo` (via `cargo-binstall`), `uv` (Python tools), `bun` (JS/TS), `npm` (Node.js), `grd` (GitHub release binaries), and `go` (Go toolchain)
- **Declarative config** — installed packages tracked in `~/.i/packages.toml`
- **Plugin drivers** — easy to extend with new package managers

## Installation

```bash
go install github.com/lucidfrontier45/i@latest
```

### Supported managers

| Manager | Requirement                                                      | Status | Command used             |
| :------ | :--------------------------------------------------------------- | :----- | :----------------------- |
| `cargo` | [`cargo-binstall`](https://github.com/cargo-bins/cargo-binstall) | ✅     | `cargo binstall`         |
| `uv`    | [`uv`](https://docs.astral.sh/uv/)                               | ✅     | `uv tool install`        |
| `bun`   | [`bun`](https://bun.sh)                                          | ✅     | `bun i -g`               |
| `npm`   | [`npm`](https://docs.npmjs.com/)                                 | ✅     | `npm i -g`               |
| `grd`   | [`grd`](https://github.com/lucidfrontier45/grd)                  | ✅     | `grd`                    |
| `go`    | [Go toolchain](https://go.dev/dl/)                               | ✅     | `go install pkg@version` |

## Usage

### Add a package

```bash
i add <package> --manager <manager> [--version <version>] [--alias <alias>]
```

Install a package and register it in the config.

```bash
i add starship --manager cargo
i add ruff --manager uv --version 0.11.0
i add lucidfrontier45/grd --manager grd
i add golang.org/x/tools/gopls --manager go
i add typescript --manager npm --version 5.6.3
```

Use `--alias` (short `-a`) to register a package under a shorter or
friendlier name. The alias becomes the key used by `upgrade`, `remove`,
`list`, and `sync`; the full package name is what gets passed to the
package manager. If `--alias` is omitted, the package name is used as the
alias:

```bash
i add @user/package --manager bun --alias mypkg
i upgrade mypkg
i remove mypkg
```

Some managers support features/extras using bracket syntax:

```bash
i add "pandas[performance,plot]" --manager uv
```

#### uv-specific options

Pass these flags to `i add` when using `--manager uv`:

| Flag      | Description                                                 |
| :-------- | :---------------------------------------------------------- |
| `--with`  | Extra package(s) to include; repeatable or comma-separated  |

```bash
# Both forms are equivalent:
i add pytest --manager uv --with httpx,mypy
i add pytest --manager uv --with httpx --with mypy

# With version constraints:
i add pytest --manager uv --with httpx==0.28.0 --with mypy==1.15.0
```

The `--with` flag accepts `name` or `name==version`, is repeatable, and also
accepts comma-separated values in a single flag. It maps to `uv tool install --with <pkg>`.

#### grd-specific options

Pass these flags to `i add` when using `--manager grd`:

| Flag            | Description                                      |
| :-------------- | :----------------------------------------------- |
| `--destination` | Destination directory for the binary             |
| `--bin-name`    | Override the binary name after download          |
| `--exclude`     | Comma-separated asset-name substrings to exclude |

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

| Flag              | Description                                   |
| :---------------- | :-------------------------------------------- |
| `--verbose`, `-v` | Print every underlying command before it runs |

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

### Get grd

```bash
i get-grd [--version <version>]
```

Download or upgrade `grd` itself from GitHub releases. If `grd` is not already on
PATH, it is installed next to the `i` binary. If it is already installed and the
current version is older than the requested one, it is upgraded in place. The
default version is `0.9.1`.

```bash
i get-grd                  # install/upgrade to 0.9.1
i get-grd --version 0.9.0  # install/upgrade to a specific version
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

## How it works

Each package manager implements the `types.Driver` interface (`internal/types/package.go`). The CLI delegates install/upgrade/remove operations to the appropriate driver. Package metadata is persisted in `~/.i/packages.toml`.

```toml
[index]
myalias = "@user/package"

[packages."starship"]
manager = "cargo"
version = "1.22.1"

[packages."ruff"]
manager = "uv"
version = "0.11.0"

[packages."@user/package"]
manager = "bun"
version = "1.0.0"
```

## Extending

To add a new package manager, implement the `types.Driver` interface and register it in an `init()` function inside the `manager` package. See `internal/manager/cargo.go` for a reference implementation.

## License

MIT
