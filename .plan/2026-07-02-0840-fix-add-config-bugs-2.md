# Fix `i` CLI — full bug sweep

> Supersedes `2026-07-02-0840-fix-add-config-bugs.md` (scope expanded from 4 to
> 8 code bugs per decision). Original preserved per skill no-overwrite rule.

## Goal

Fix all 8 confirmed code-level defects in the `i` CLI (config-state corruption,
silent option loss, dead code, interface inconsistency, version-tag mismatch,
no mutual exclusion) in a single commit/PR. **No test additions this pass.**

## Background

Bug audit via CodeGraph + runtime reproduction (BurntSushi v1.6.0 round-trip
test) + live GitHub release metadata. The `self-upgrade` download/checksum path
and `get-grd` default-version path were audited and are **correct** — untouched.

Defects split: B1–B4 (config-state correctness, originally confirmed) +
B5–B8 (lower-severity cleanup/robustness, now in scope).

### B1 — uv `--with` dropped after config reload  [HIGH]
`with` is the only array option. In-memory `[]string` → TOML → reloaded as
`[]interface{}`. `.([]string)` assertion in `internal/manager/uv.go:41`,`:59`
fails → `--with` companion packages vanish on every op after reload.
Reproduced:
```
with -> []interface {} ...  with .([]string) FAILED
```

### B2 — `add` on existing package ignores manager change → drift  [MEDIUM]
`cmd/add.go` update branch: `merged := existing` updates version/features/options
but never `merged.Manager`; no validation rejects a change. Re-adding an
npm-managed pkg with `--manager bun` installs via bun, config keeps
`manager = "npm"` → all later ops hit wrong driver.

### B3 — `add` update replaces whole Options map → siblings wiped  [MEDIUM]
`cmd/add.go` (~L127) `merged.Options = options` is full replace. version/features
merge field-by-field; Options does not. `--destination /a` then `--bin-name x`
silently loses `/a` (and `exclude`/`with`).

### B4 — `upgrade --manager` flag declared but never read  [LOW-MED]
`cmd/upgrade.go:136` registers `--manager`; `RunE` discards `cmd`; `runUpgrade`
never reads it. `i upgrade --manager go` upgrades ALL managers' packages.

### B5 — `go` `InstalledVersion` ignores `bin-name` option  [LOW]
`internal/manager/go.go:68` uses `filepath.Base(pkg)`; `Remove` (`:50`) honors
`bin-name` via `binNameFromSpec`. Mismatch: version refresh silently no-ops for
renamed go binaries (caller masks the error). Root cause: interface signature
`InstalledVersion(ctx, pkg string)` cannot see options.

### B6 — `get-grd --version vX.Y.Z` (explicit `v`) → 404  [LOW]
`cmd/get_grd.go` `downloadGrd` passes `version` un-normalized to `ReleaseByTag`.
grd tags are unprefixed (`0.11.0`), so a user-supplied `v0.11.0` →
`/releases/tags/v0.11.0` → 404. Default `0.11.0` and `--latest` work; only
explicit-`v` breaks. semver compare path is fine (uses `normalizeV`).

### B7 — dead code: `manager.Detect()` / `All()` / driver `Detect()`  [LOW]
`internal/manager/registry.go` `Detect()` and `All()` have zero callers
(verified via `codegraph callers` + grep). Each driver's `Detect() bool` only
exists to satisfy the interface and is invoked by nothing. Commands always
trust config/`--manager`; availability is never probed.

### B8 — no mutual exclusion on `packages.toml`  [LOW]
`config.Read` → mutate → `Write` is non-atomic across invocations. Two
concurrent `i` runs (e.g. terminal A `add`, terminal B `sync`) interleave
Read/Write → lost updates. Per-`Write` is atomic (temp+rename), the RMW is not.

## Approach

Decisions: B1 → normalize on `config.Read`. B2 → **reject** manager change.
B4 → **remove** dead flag. B5 → widen interface to pass `PackageSpec`.
B7 → **remove** dead code. B8 → portable `O_EXCL` advisory lockfile. Single
commit/PR, no tests.

Order = isolated one-liners first, invasive interface sweep + locking last.

1. **B1 — normalize option types on read** (`internal/config/toml.go`)
   - After `toml.Unmarshal`, per entry coerce `Options["with"]` from `[]any` →
     `[]string` (drop non-string elements) via private `normalizeOptions`.
   - Idempotent; scalars (`force`/`destination`/`bin-name`/`exclude`) untouched.
   - `uv.go` assertions now pass post-reload. No driver change.

2. **B2 — reject manager change** (`cmd/add.go`)
   - In update branch, before any mutation, guard:
     `if mgr != existing.Manager → error "package %q is managed by %q; run 'i remove %s' first to switch managers"`.
   - Place after alias-conflict checks, before index mutation. `merged.Manager`
     stays == existing.

3. **B3 — merge Options per-key** (`cmd/add.go`)
   - Replace `merged.Options = options` with:
     `if merged.Options == nil { merged.Options = map[string]any{} }; for k,v := range options { merged.Options[k] = v }`.
   - Keep `len(options) > 0` guard so no-op adds don't set `changed`.

4. **B4 — remove dead `--manager` from upgrade** (`cmd/upgrade.go`)
   - Delete `upgradeCmd.Flags().String("manager", ...)` (L136). Nothing else
     references it.

5. **B6 — strip `v` prefix before grd tag lookup** (`cmd/get_grd.go`)
   - In `downloadGrd`, `tag := strings.TrimPrefix(version, "v")`; pass `tag` to
     `ReleaseByTag` and the final print. `runGetGrd`'s `semver.Compare` path
     (uses `normalizeV`) unaffected.

6. **Interface + driver sweep (B5 + B7 combined — both touch the interface)**
   - `internal/types/package.go`:
     - Change `InstalledVersion(ctx context.Context, pkg string)` →
       `InstalledVersion(ctx context.Context, spec PackageSpec)`.
     - Remove `Detect() bool` from the interface.
   - 6 drivers (`bun/cargo/go/grd/npm/uv.go`):
     - Update `InstalledVersion` signature; in `go.go` replace
       `filepath.Base(pkg)` with `binNameFromSpec(spec)` (the actual B5 fix).
       Other drivers ignore the extra fields (already take `pkg` from
       `spec.Name`).
     - Delete each `Detect()` method body (6 files).
   - `internal/manager/registry.go`: delete `Detect()` and `All()`; keep
     `Register`, `Lookup`, `drivers`.
   - Callers (`add.go` ×2, `sync.go` ×1, `upgrade.go` ×2): change
     `drv.InstalledVersion(ctx, string(name))` → `drv.InstalledVersion(ctx, spec)`.
     `spec` is already in scope at every call site.
   - `go vet`/build will catch any missed site.

7. **B8 — advisory lockfile around RMW** (`internal/config` + commands)
   - Add `config.WithLock(fn func() error) error` using an exclusive lockfile
     `~/.i/.packages.toml.lock`, created with `os.OpenFile(..., O_CREATE|O_EXCL)`.
   - Stale handling: write PID into the lockfile; on `O_EXCL` failure, read PID,
     if process not alive (`os.FindProcess`+`Signal(0)`) steal the lock; else
     bounded retry (e.g. 30 × 200ms) then error `"another i instance is running
     (PID %d); waiting…"`.
   - Wrap the full Read..Write span in `runAdd`, `runRemove`, `runSync`,
     `runUpgrade`. Note: sync/upgrade hold the lock across all installs (long,
     but concurrent sync is rare and serialization is the point).
   - `list` is read-only → no lock needed (tolerates a stale read).

8. **Validate & lint** (per AGENTS.md)
   - `go build ./...`, `go vet ./...`, `go mod tidy`.
   - `golangci-lint run --fix` then `golangci-lint fmt`.

9. **Single commit** (conventional-commits `fix:`), body bullets B1–B8 with
   before/after behavior. Scope `config`/`cmd`/`manager` as fits.

## Trade-offs

- **B1 normalize-on-Read vs tolerant `uv.go` helper.** Chose Read: single point,
  fixes any future array option, drivers stay simple. Helper would scatter the
  workaround and leave `[]any` in config for later consumers.
- **B2 reject vs allow-and-switch.** Chose reject: matches declarative-config
  model, clear remediation, no silent cross-manager migration.
- **B4 remove vs implement filtering.** Chose remove: smallest correct change,
  kills the footgun. Real per-manager upgrade needs detection + partial-failure
  UX — separate feature.
- **B5 widen interface vs defer/document.** Chose widen: only way `go`
  `InstalledVersion` can honor `bin-name`. Cost: mechanical change across
  interface + 6 drivers + 5 callers. Rejected defer: leaves the inconsistency
  and a silent stale-version path.
- **B7 remove vs wire-into-`add` for friendly "tool not installed" errors.**
  Chose remove: dead code is dead; removal is the honest fix. Wiring in is a
  feature (better error UX) and scope creep. Flagged as future work.
- **B8 `O_EXCL` lockfile vs `golang.org/x/sys` flock.** Chose `O_EXCL`:
  portable, no new dependency, adequate for a low-contention CLI. flock would
  need a unix/windows split + new dep (`golang.org/x/sys`).
- **Single commit vs per-area commits.** Honoring prior decision (single PR),
  but flag: this diff is now large (interface widening + locking + 8 fixes).
  Reviewer may prefer splits — e.g. `fix(config)`, `fix(manager)`,
  `feat(config): locking`. See open questions.

## Open questions

- **B9 — test coverage.** "Include those too" conflicts with the standing
  "no tests for now" decision. **B9 kept EXCLUDED**; confirm if you now want
  tests added (highest-value targets: B1 round-trip, B2/B3 update branch, B8
  lock contention).
- **B8 lock granularity & policy.** Hold across full sync (simple, long) or
  only around Read+Write pairs (complex, still long)? Stale-lock timeout
  value? PID-steal on dead owner — acceptable, or fail-closed? Defaults above
  are proposals.
- **Single commit vs split.** Diff is sizable. Keep single `fix:` commit, or
  split into logical commits (config / manager-interface / locking)?
- **B4 replacement.** Remove outright (current plan), or keep `--manager` and
  actually implement filtering in `runUpgrade`? Default = remove.
- **B2 error wording.** Draft prescriptive; confirm tone/wording.

## Next step

Begin with **B1** (`internal/config/toml.go` normalize) — root cause, isolated.
Then B2→B3 (`cmd/add.go`), B4 (`cmd/upgrade.go`), B6 (`cmd/get_grd.go`); then
the combined **interface sweep (B5+B7)**; then **B8 locking**; then
build/lint/tidy; then commit. No code written until you confirm.
