# TODO: 2026-07-02-0840-fix-add-config-bugs-2.md

**PR Created:** https://github.com/lucidfrontier45/i/pull/20

## Phase 1: B1 — normalize option types on read (internal/config/toml.go)
- [x] Add private `normalizeOptions` function to coerce `Options["with"]` from `[]any` → `[]string` after `toml.Unmarshal`
- [x] Test: ensure uv assertions pass post-reload, scalars untouched

## Phase 2: B2 — reject manager change (cmd/add.go)
- [x] Add guard in update branch: if `mgr != existing.Manager` return error with prescriptive message
- [x] Place guard after alias-conflict checks, before index mutation

## Phase 3: B3 — merge Options per-key (cmd/add.go)
- [x] Replace `merged.Options = options` with per-key merge, guard with `len(options) > 0`
- [x] Keep `changed` logic correct for no-op adds

## Phase 4: B4 — remove dead `--manager` from upgrade (cmd/upgrade.go)
- [x] Delete `upgradeCmd.Flags().String("manager", ...)` at L136

## Phase 5: B6 — strip `v` prefix before grd tag lookup (cmd/get_grd.go)
- [x] In `downloadGrd`, add `tag := strings.TrimPrefix(version, "v")` and pass `tag` to `ReleaseByTag`

## Phase 6: Interface + driver sweep (B5 + B7 combined)
- [x] Change `internal/types/package.go`: `InstalledVersion` signature to `PackageSpec`, remove `Detect() bool`
- [x] Update 6 drivers (`bun/cargo/go/grd/npm/uv.go`): `InstalledVersion` signature, delete `Detect()` bodies
- [x] In `go.go`: replace `filepath.Base(pkg)` with `binNameFromSpec(spec)` for B5 fix
- [x] Delete `internal/manager/registry.go`: `Detect()` and `All()` methods
- [x] Update callers (add.go ×2, sync.go ×1, upgrade.go ×2): pass `spec` instead of `string(name)`
- [x] Verify: `go build ./...` and `go vet ./...` catch any missed sites

## Phase 7: B8 — advisory lockfile (DEFERRED out of this PR)
- Removed: WithLock API landed with zero callers and a Windows-broken stale
  check (syscall.Signal(0) unsupported -> lock always stolen). Re-introduce in
  a follow-up PR with a cross-platform liveness check and actual wiring into
  runAdd/runSync/runUpgrade/runRemove. Design preserved in .plan §7.

## Phase 8: Validate & lint (per AGENTS.md)
- [x] Run `go build ./...`, `go vet ./...`, `go mod tidy`
- [x] Run `golangci-lint run --fix` then `golangci-lint fmt`

## Phase 9: Single commit
- [x] Create conventional-commits `fix:` commit with B1–B8 bullets and before/after behavior