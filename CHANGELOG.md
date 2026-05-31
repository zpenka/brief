# Changelog

## v0.1.0 — 2026-05-31

Initial release.

### Features

- Snapshot branch, recent commits, working tree status, and TODO/FIXME/HACK/NOTE/XXX markers from the unstaged diff
- Capture staged diff separately (`git diff --cached`) and surface TODO markers from staged changes in a dedicated section
- Run `go test -json -race ./...` via `--tests` and include pass/fail counts; failed tests include assertion messages and file:line details
- JSON output mode (`--json`) for scripting
- `--version` flag
