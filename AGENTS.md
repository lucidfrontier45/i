# AGENTS.md

This project uses Go. AI agents must use **golangci-lint v2** for all code quality and formatting tasks.

## 🛑 Strict Policy: No Commits
**Do not perform any git commits, pushes, or version control actions.** Your responsibility ends at preparing the code changes and ensuring they pass all local checks. The human user will handle all commits.

## Setup Commands
- **Initialize:** `go mod download`
- **Maintenance:** `go mod tidy`

## Code Quality & Formatting
Use `golangci-lint` exclusively. Do not use standalone formatters.

| Task        | Command                   | Description                              |
| :---------- | :------------------------ | :--------------------------------------- |
| **Check**   | `golangci-lint run`       | Runs all enabled linters and formatters. |
| **Autofix** | `golangci-lint run --fix` | Applies linter fixes and formats code.   |
| **Format**  | `golangci-lint fmt`       | Runs configured v2 formatters only.      |

## Project Structure
Detect the existing layout or request guidance for new projects:

1. **Existing Projects:** Detect if the project uses a **Flat Layout** (logic in root) or **Standard Layout** (using `internal/`, `pkg/`, and `cmd/`). Follow the established pattern strictly.
2. **New Projects:** **Do not assume a structure.** Ask the user: "Would you prefer a Flat Layout or a Standard Go Project Layout (`internal/pkg/cmd`)?" before creating directories.

## Build & Test
- **Build:** `go build ./...`
- **Test:** `go test ./...`
- **Race Check:** `go test -race ./...`

## Development Guidelines
- **Error Handling:** Handle all errors explicitly. Do not use `_` without a documented reason.
- **Documentation:** Every exported identifier must have a comment starting with its name.
- **V2 Config:** Ensure `.golangci.yml` uses `version: "2"`.
- **LSP Warnings:** Ignore unused imports/warnings from LSP; they will be resolved by the linter's autofix and formatter steps.
- **Loop Style:** Prefer `for i := range N` over `for i := 0; i<N; i++` for index-based iteration loops.

## Preparation for Human Review
- Ensure `go mod tidy` and `golangci-lint fmt` have been run.
- Summarize the changes made and list any new dependencies added so the human can review before committing.