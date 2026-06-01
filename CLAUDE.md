# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./cmd/brief

# Test (all, with race detector)
go test -race ./...

# Test (single)
go test -run TestParseBranch ./...

# Vet + format check
go vet ./...
gofmt -l .   # should return nothing

# Run
go run ./cmd/brief                        # current directory
go run ./cmd/brief --dir /path/to/repo    # explicit repo
go run ./cmd/brief --base main            # compare against base branch
go run ./cmd/brief --tree                 # include repo file tree
go run ./cmd/brief --tokens 4000          # cap at ~4k tokens
go run ./cmd/brief --tests                # include test run
go run ./cmd/brief --json                 # JSON output
```

## Architecture

`brief` is a CLI tool that generates a compact, markdown-formatted snapshot of a git repository's current state, intended to be handed to Claude (or any LLM) as session context — replacing the 4–6 tool calls Claude typically makes at session start.

**Package layout (all in root package `brief`):**

- `cmd/brief/main.go` — thin entry point; calls `brief.Run()` and exits
- `brief.go` — `Config`, `Brief` struct, `Run()` (flag parsing), `Collect()` (data gathering), `printText()`, `printJSON()`
- `git.go` — `currentBranch()`, `recentCommits()`, `workingStatus()`, `diffOutput()`, `commitsAhead()`, `baseDiff()` via `git` exec; `parseBranch()`, `parseCommits()`, `parseStatus()` are pure functions tested directly
- `todos.go` — `todosInDiff()` scans unified diff output for TODO/FIXME/HACK/NOTE/XXX markers on added lines; `parseHunkStart()` extracts new-file line numbers from `@@` headers
- `tests.go` — `runTests()` shells out to `go test -json -race ./...`; `parseTestEvents()` decodes the JSON event stream
- `tree.go` — `repoFiles()` collects tracked + untracked non-ignored paths via `git ls-files`; `buildTree()` formats them as an indented directory tree up to a configurable depth
- `truncate.go` — `estimateTokens()` (len/4 heuristic), `applyBudget()` trims `BranchDiff` first then `Commits` to stay within a token budget

**Data flow:** `Run` → `Collect` → `currentBranch`, `recentCommits`, `workingStatus`, `diffOutput` → `todosInDiff` on diff → optional `commitsAhead`+`baseDiff` (if `--base`) → optional `repoFiles`+`buildTree` (if `--tree`) → optional `applyBudget` (if `--tokens`) → optional `runTests` → `printText` or `printJSON`

**Output format:** markdown, designed for direct consumption by LLMs. Sections: branch, recent commits, working tree status, todos in diff (only if any), branch commits vs base + diff + todos (only if `--base`), repo tree (only if `--tree`), tests (only if `--tests`).

**No external dependencies** — stdlib only.

**Testing approach:** pure parse functions are tested with fixture strings; integration tests (`TestCurrentBranch`, `TestCollect_*`) use `tempGitRepo` (defined in `git_test.go`) to create a real git repo in `t.TempDir()`.
