# Plan: Add `version` Subcommand

**Issue:** #6 — "add version subcommand"
**Branch:** `feat/version-subcommand`

---

## Summary

Add a `version` subcommand that prints the tool's version, commit hash, and build date. Wire ldflags in `.goreleaser.yaml` so releases carry real version info, and the dev build defaults to `"dev"`.

---

## Files to Create

### `cmd/version.go`

New CLI subcommand following the same pattern as `list.go`, `remove.go`, etc.

```go
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)
```

- Define package-level vars: `var (Version = "dev"; Commit = ""; Date = "")`.
- Create `versionCmd` with `Use: "version"`, `Short: "Print version information"`, `Args: cobra.NoArgs`.
- `RunE` prints:
  ```
  i version dev
  commit <commit>
  built at <date>
  go version <runtime.Version()>
  os/arch <runtime.GOOS>/<runtime.GOARCH>
  ```
- Register in `init()`: `rootCmd.AddCommand(versionCmd)`.

---

## Files to Modify

### `.goreleaser.yaml`

Add an `ldflags` section under `builds`:

```yaml
ldflags:
  - -s -w
  - -X github.com/lucidfrontier45/i/cmd.Version={{.Version}}
  - -X github.com/lucidfrontier45/i/cmd.Commit={{.Commit}}
  - -X github.com/lucidfrontier45/i/cmd.Date={{.Date}}
```

---

## Implementation Order

1. Create `cmd/version.go` with vars and subcommand.
2. Add ldflags to `.goreleaser.yaml`.
3. `go mod tidy` (if needed).
4. Run `golangci-lint fmt` and `golangci-lint run --fix`.
5. `go build ./...` and `go test ./...`.
6. Manual verification: `go build -o i.exe . && ./i.exe version`.
