# AGENTS.md

## Project overview

`stackit-acl` is a Go CLI tool that manages the caller's external IP address in the ACLs of STACKIT cloud services (PostgreSQL Flex, MongoDB Flex, Redis, SKE, etc.). It fetches the external IP, diffs it against a resource's current ACL, shows the planned change, asks for confirmation, and applies it by shelling out to the `stackit` CLI.

- Module: `github.com/hiqs-gmbh/stackit-acl` (Go 1.27)
- The `stackit` CLI is a runtime dependency, invoked via `exec.Command` — there is no STACKIT API client in this codebase.

## Commands

- Build: `make build` (writes `bin/stackit-acl`)
- Test: `make test`
- Lint: `make lint` (golangci-lint; must pass with 0 issues)
- Format check: `gofmt -l .` (must print nothing)
- All three: `make lint build test` — run this before finishing any change.

## Repository layout

- `main.go` — entry point, delegates to `internal/cmd`.
- `internal/cmd/` — cobra commands and CLI glue.
  - `root.go` — root command, flags, arg validation/completion, and `runACLAction` (the plan → confirm → apply flow, supports multiple resource IDs / bulk mode).
  - `add.go`, `remove.go` — thin command definitions, both call `runACLAction`.
  - `output.go` — custom `slog` handler with colored output and the `logStep`/`logInfo`/`logSuccess`/`logWarn`/`logPrompt`/`logRaw` helpers.
- `internal/services/service.go` — registry of supported STACKIT services (`ServiceConfig`: JSON paths, ACL type, update strategy).
- `internal/acl/` — pure functions: CIDR conversion, ACL extraction/insertion from JSON payloads.
- `internal/ip/` — external IP fetching (`https://ifconfig.schwarz`).
- `internal/stackit/client.go` — wrapper around the `stackit` CLI (`runRaw` builds and executes the command).

## Code style

- Standard Go formatting (`gofmt`), no comments in code, small focused functions.
- Commit messages: single line, imperative mood, capitalized, no prefix (e.g. `Add bulk update support for multiple resource IDs`). One logical change per commit.
- Commit yourself after completing a significant change (feature, bugfix, config change); split unrelated changes into separate commits. Verify `make lint build test` passes before committing.

## Logging conventions (important)

- **All console output goes through `log/slog`** using the helpers in `internal/cmd/output.go`. Never use `fmt.Print*`/`fmt.Fprint*`/`print`/`println` — this is enforced by the `forbidigo` linter in `.golangci.yaml`.
- Output kinds are passed as a `kind` slog attribute: `step` (→), `success` (✓), `warn` (⚠), `error` (✗), `prompt` (bold, no trailing newline), `raw` (verbatim, no prefix). Errors returned from `Execute()` are logged by `main.go`.
- `fmt.Errorf`, `fmt.Sprintf`, and `fmt.Fprintf` into a `strings.Builder` are fine for building strings — but prefer plain concatenation where trivial, since `staticcheck` (QF1012) flags `WriteString(fmt.Sprintf(...))`.
- `errcheck` is enabled: explicitly discard ignored errors (`_, _ = w.Write(...)` or `defer func() { _ = f.Close() }()`).

## Architecture notes

- Adding a new STACKIT service = adding one entry to the `registry` map in `internal/services/service.go`. Two ACL shapes exist (`ACLArray`, `ACLCommaString`) and two update strategies (`FlagStrategy` passes `--acl` flags; `PayloadStrategy` uses `generate-payload` + `acl.SetACLs` + `update --payload @file`, currently only SKE).
- `stackit.Client` appends global args (`-p`, `--verbosity error`, `--region`) exactly once, centrally in `run()` — do not re-append them in individual methods.
- `runACLAction` flow: fetch IP once → for each resource, fetch current state and skip no-ops → show all planned changes → one single confirmation for the whole batch → apply all (continue on per-resource failure, then return a summary error).
- Update README.md (supported services table, examples) whenever commands, flags, or services change.

## Testing

- Unit tests exist only for pure packages (`internal/acl`, `internal/ip`, `internal/services`); `internal/cmd` and `internal/stackit` have none.
- Manual end-to-end verification: create a stub `stackit` script on `PATH` that handles `<service> <group> describe|update|generate-payload`, then run the built binary against fake resource IDs. Cover: happy path, confirmation prompt (pipe `y`/`n` to stdin), bulk with mixed no-ops, and a failing update to check the error summary and exit code.

## Releases

- Version lives in `const Version` in `internal/cmd/root.go`.
- Release flow: bump the constant → commit as `Bump version to X.Y.Z` → create a **lightweight** tag `vX.Y.Z` (repo convention) → `git push origin main vX.Y.Z`.
- Note: `git push --follow-tags` does **not** push lightweight tags — always push tags explicitly.
