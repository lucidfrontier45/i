# Add `i alias` subcommand

Created: 2026-10-06 21:12
Status: draft, awaiting implementation

## Goal

Add a config-only `i alias set|list|remove` command group so aliases can be created,
inspected, and dropped for already-registered packages without reinstalling.

## Background

`i` persists a declarative manifest at `~/.i/packages.toml` with two maps
(`internal/config/toml.go:21`):

- `Index` — `PackageAlias → PackageName`, the alias indirection
- `Packages` — `PackageName → PackageEntry`, the canonical registry

Today an alias can **only** be created at install time via `i add -a <alias> <pkg>`
(`cmd/add.go:242`), and changing one requires re-running `add`, which reinstalls the
package and re-queries `InstalledVersion` (`cmd/add.go:146-155`). `i list` renders
aliases in a comma-joined column (`cmd/list.go:58`), so it already tolerates several
per package. There is no read/modify path for `Index` outside `add` and `remove`.

## Decisions

Settled during planning:

| Decision | Choice |
| :------- | :----- |
| Command surface | Full CRUD: `i alias set \| list \| remove` (flat, not a nested group) |
| Side effects | Config-only; no `manager.Lookup`, no driver calls |
| Cardinality | Many aliases may point to one package |
| Re-pointing an existing alias | Refused unless `--force` |
| Removal | Keyed strictly by alias; a package name is not accepted |
| Validation | Self-contained in `cmd/alias.go`; no shared helpers with `add.go` |
| Tests | None added (repo currently has zero `*_test.go` files) |

## Approach

1. **`cmd/alias.go` (new)** — parent `alias` command with
   `Short: "Manage package aliases"`, no `RunE` (prints help when invoked bare),
   registered via `rootCmd.AddCommand` in `init()`.

2. **`i alias set <alias> <package> [--force]`** — resolve `<package>` through
   `cfg.ResolveName` so either the full name or an existing alias may be passed;
   verify the target exists in `cfg.Packages`; reject the alias if it collides with a
   *different* registered package name; reject alias == target name as redundant.

   **Re-pointing is refused by default.** If `cfg.Index[alias]` already maps to a
   different package, return an error naming the current target and instructing the
   user to pass `--force`. With `--force`, overwrite and print a line recording the
   previous target. Mapping the alias to the package it already points at is
   idempotent — a no-op success, no `--force` required. All paths end in
   `config.Write`. Pure `config.Read`/`config.Write` — no driver calls.

3. **`i alias list`** — tabwriter over `cfg.Index` sorted by alias, columns
   `Alias` / `Package`; print `no aliases configured` when empty (mirrors
   `cmd/list.go:22`). Read-only, no write.

4. **`i alias remove <alias>`** — keyed strictly by alias. Resolve nothing, accept no
   package name: delete the key from `Index`, error `alias %q not found` when absent,
   write otherwise. The package and its install are left untouched; dropping a
   package's aliases remains the job of `i remove <pkg>`, which already sweeps
   `cfg.Index` (`cmd/remove.go:40-44`).

5. **`cmd/add.go:106-113`** — delete the loop that strips all prior aliases pointing at
   `pkg` before assigning `aliasFlag`, so `add -a` becomes additive instead of
   replacing. Keep the existing collision guard at `cmd/add.go:76-78` (alias must not
   shadow a registered package name) and `cmd/add.go:79-81` (already-mapped-to-same-
   package stays a no-op).

6. **Docs** — README: add an `i alias` section next to the existing `--alias` prose
   (README.md:51-58) and the `[index]` TOML example at README.md:191.
   `docs/architecture.md`: add `alias.go` to the `cmd/` tree listing and record the
   many-alias-per-package decision under Key Design Decisions (it currently describes
   the alias index as a simple indirection).

7. **Verify** — `go mod tidy`, `golangci-lint fmt`, `golangci-lint run --fix`,
   `go build ./...`. No tests added, per the decisions above.

## Trade-offs

- **Config-only, no driver calls.** Alias is a pure lookup key; nothing in the
  installed artifact depends on it. Reinstalling to rename a string would be slow,
  require the manager tool on PATH, and could fail for reasons unrelated to the alias.
- **`--force` to re-point.** Costs a flag on a rare operation, but silently moving a
  user's shorthand onto a different package is exactly the kind of surprise that makes
  someone lose track of what `i rm foo` will actually remove. Failing loudly costs one
  extra command the first time and nothing after.
- **Alias-keyed removal only.** Simpler and unambiguous — the argument's meaning never
  depends on lookup order. The cost is no bulk detach, which `i remove <pkg>` already
  covers.
- **Flat `alias set|list|remove` over nested subcommand groups.** Fewer levels to type,
  but these are bare words that could collide with future verbs; a nested
  `i alias add/rm/list` group would scale better. Kept flat because `set`/`remove`
  reads more clearly than `add`/`rm` at this size.
- **Validation duplicated in `alias.go` rather than extracted.** Keeps the diff to two
  files and honors the repo's no-tests stance, but the collision rules now exist in two
  places and can drift.
- **Many-to-one aliases.** More flexible and already supported by `list`'s rendering,
  at the cost of `add.go` semantics changing for existing users who relied on `-a`
  silently replacing the previous alias.

## Open questions

None outstanding.

## Next step

Create `cmd/alias.go`, adjust `cmd/add.go`, update README and architecture doc, then
run the lint + build chain.