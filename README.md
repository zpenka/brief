# brief

**brief** generates a compact, markdown-formatted snapshot of your git repository's current state — branch, recent commits, working tree status, TODO markers in your diff, base-branch comparison, and repo file tree — in one shot.

Designed to be handed to Claude (or any LLM) as session context, replacing the 4–6 tool calls it would otherwise make at the start of a session.

## Install

```bash
brew install zpenka/tap/brief
```

Or via Go:

```bash
go install github.com/zpenka/brief/cmd/brief@latest
```

## Usage

```
brief [--dir <path>] [--commits <n>] [--base <branch>] [--tree] [--tokens <n>] [--tests] [--json]
```

```bash
# Snapshot the current repo
brief

# Snapshot a different repo
brief --dir ~/code/myproject

# Compare against a base branch (commits ahead + full diff)
brief --base main

# Include the repo file tree
brief --tree

# Cap output at ~4 000 tokens (safe for most context windows)
brief --tokens 4000

# Include test results
brief --tests

# JSON output (for scripting)
brief --json
```

## Example output

```
# brief: /Users/zpenka/code/myapp

**branch:** feature/auth-refresh
**as of:** 2026-05-31 14:23

## recent commits
- `a1b2c3d` fix: token expiry not checked on refresh
- `e4f5a6b` add refresh endpoint
- `c7d8e9f` initial auth scaffold

## working tree
- modified: `internal/auth/handler.go`
- untracked: `internal/auth/handler_test.go`

## todos in recent changes
- `internal/auth/handler.go:42` **TODO**: handle concurrent refresh race
- `internal/auth/handler.go:87` **FIXME**: error message leaks token length

## tests
✓ 134 passed (4.2s)
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `--dir` | `.` | Git repository directory |
| `--commits` | `5` | Number of recent commits |
| `--base` | — | Base branch to compare against (e.g. `main`). Shows commits ahead and full diff. |
| `--tree` | off | Include a condensed repo file tree |
| `--depth` | `3` | Max directory depth for `--tree` |
| `--tokens` | `0` | Token budget for output (`0` = unlimited). Truncates diff first, then trims commits. |
| `--tests` | off | Run `go test -race ./...` and include result |
| `--json` | off | Output JSON instead of markdown |

## Use with Claude

Pipe the output directly into a Claude session as context:

```bash
brief > context.md
# then reference context.md when starting your session
```

Or, with tools that support stdin context injection:

```bash
brief | claude --context -
```
