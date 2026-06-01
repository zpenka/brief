package brief

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollect_Basic(t *testing.T) {
	dir := tempGitRepo(t)

	b, err := Collect(Config{Dir: dir, NumCommits: 5})
	if err != nil {
		t.Fatal(err)
	}
	if b.Branch == "" {
		t.Error("expected non-empty branch")
	}
	if len(b.Commits) == 0 {
		t.Error("expected at least one commit")
	}
	if b.Dir == "" {
		t.Error("expected non-empty Dir")
	}
}

func TestCollect_DirtyStatus(t *testing.T) {
	dir := tempGitRepo(t)

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n// changed\n"), 0644); err != nil {
		t.Fatal(err)
	}

	b, err := Collect(Config{Dir: dir, NumCommits: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Status) == 0 {
		t.Error("expected dirty status")
	}
}

func TestCollect_TODOsInDiff(t *testing.T) {
	dir := tempGitRepo(t)

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n// TODO: do something\n"), 0644); err != nil {
		t.Fatal(err)
	}

	b, err := Collect(Config{Dir: dir, NumCommits: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.TODOs) == 0 {
		t.Error("expected TODO item in diff")
	}
	if b.TODOs[0].Kind != "TODO" {
		t.Errorf("Kind = %q, want TODO", b.TODOs[0].Kind)
	}
}

func TestCollect_StagedTODOs(t *testing.T) {
	dir := tempGitRepo(t)

	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n// TODO: staged work\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("git", "add", "main.go")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}

	b, err := Collect(Config{Dir: dir, NumCommits: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.StagedTODOs) == 0 {
		t.Error("expected StagedTODOs from staged diff")
	}
	if b.StagedTODOs[0].Kind != "TODO" {
		t.Errorf("Kind = %q, want TODO", b.StagedTODOs[0].Kind)
	}
	if b.StagedTODOs[0].Text != "staged work" {
		t.Errorf("Text = %q, want %q", b.StagedTODOs[0].Text, "staged work")
	}
}

func TestCollect_UnstagedTODOs_NotInStaged(t *testing.T) {
	dir := tempGitRepo(t)

	// Write TODO but do NOT stage it
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n// TODO: unstaged work\n"), 0644); err != nil {
		t.Fatal(err)
	}

	b, err := Collect(Config{Dir: dir, NumCommits: 5})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.StagedTODOs) != 0 {
		t.Errorf("expected no StagedTODOs for unstaged change, got %v", b.StagedTODOs)
	}
	// TODOs in unstaged diff (via git diff HEAD) should still appear
	if len(b.TODOs) == 0 {
		t.Error("expected TODOs from unstaged diff")
	}
}

func TestCollect_MultipleCommits(t *testing.T) {
	dir := tempGitRepo(t)

	for i := 0; i < 3; i++ {
		content := []byte(fmt.Sprintf("package main\n// iteration %d\n", i))
		if err := os.WriteFile(filepath.Join(dir, "main.go"), content, 0644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("git", "commit", "-am", fmt.Sprintf("commit %d", i))
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git commit: %v\n%s", err, out)
		}
	}

	b, err := Collect(Config{Dir: dir, NumCommits: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Commits) != 2 {
		t.Errorf("len(Commits) = %d, want 2", len(b.Commits))
	}
}

func TestPrintText_ContainsSections(t *testing.T) {
	dir := tempGitRepo(t)
	b, err := Collect(Config{Dir: dir, NumCommits: 5})
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := printText(&buf, b); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	for _, want := range []string{"branch", "recent commits", "working tree"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintText_TODOsAndTests(t *testing.T) {
	b := &Brief{
		Dir:    "/repo",
		Branch: "main",
		TODOs: []TODOItem{
			{File: "auth.go", Line: 42, Kind: "FIXME", Text: "handle race"},
		},
		Tests: &TestResult{
			Passed:  true,
			Count:   12,
			Elapsed: 1200 * 1000000, // 1.2s in nanoseconds
		},
	}

	var buf bytes.Buffer
	if err := printText(&buf, b); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	for _, want := range []string{"todos in recent changes", "auth.go:42", "FIXME", "12 passed"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintText_StagedTODOs(t *testing.T) {
	b := &Brief{
		Dir:    "/repo",
		Branch: "main",
		StagedTODOs: []TODOItem{
			{File: "api.go", Line: 7, Kind: "TODO", Text: "wire up handler"},
		},
	}

	var buf bytes.Buffer
	if err := printText(&buf, b); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	for _, want := range []string{"todos in staged changes", "api.go:7", "TODO", "wire up handler"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintText_FailedTests(t *testing.T) {
	b := &Brief{
		Dir:    "/repo",
		Branch: "main",
		Tests: &TestResult{
			Passed: false,
			Failed: []string{"TestFoo", "TestBar"},
		},
	}

	var buf bytes.Buffer
	if err := printText(&buf, b); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	for _, want := range []string{"failed", "TestFoo", "TestBar"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintText_FailedTestsWithDetails(t *testing.T) {
	b := &Brief{
		Dir:    "/repo",
		Branch: "main",
		Tests: &TestResult{
			Passed: false,
			Failed: []string{"TestFoo"},
			Details: []FailureDetail{
				{Test: "TestFoo", Messages: []string{"    foo_test.go:42: assertion failed\n"}},
			},
		},
	}

	var buf bytes.Buffer
	if err := printText(&buf, b); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	for _, want := range []string{"TestFoo", "foo_test.go:42", "assertion failed"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintJSON(t *testing.T) {
	dir := tempGitRepo(t)
	b, err := Collect(Config{Dir: dir, NumCommits: 5})
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := printJSON(&buf, b); err != nil {
		t.Fatal(err)
	}

	var got map[string]any
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}
	for _, key := range []string{"dir", "branch", "commits", "status", "at"} {
		if _, ok := got[key]; !ok {
			t.Errorf("JSON missing key %q", key)
		}
	}
}

func TestCollect_WithBase(t *testing.T) {
	dir := tempGitRepo(t)
	base, err := currentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}

	mustGit(t, dir, "checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(dir, "feature.go"), []byte("package main\n// feature work\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "feature work")

	b, err := Collect(Config{Dir: dir, NumCommits: 5, Base: base})
	if err != nil {
		t.Fatal(err)
	}
	if b.BaseBranch != base {
		t.Errorf("BaseBranch = %q, want %q", b.BaseBranch, base)
	}
	if len(b.BranchCommits) == 0 {
		t.Error("expected BranchCommits to be populated")
	}
	if b.BranchDiff == "" {
		t.Error("expected BranchDiff to be populated")
	}
	if !strings.Contains(b.BranchDiff, "feature work") {
		t.Errorf("BranchDiff should contain added content, got %q", b.BranchDiff)
	}
}

func TestCollect_WithBase_TODOs(t *testing.T) {
	dir := tempGitRepo(t)
	base, err := currentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}

	mustGit(t, dir, "checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(dir, "work.go"), []byte("package main\n// TODO: finish this\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "add work")

	b, err := Collect(Config{Dir: dir, NumCommits: 5, Base: base})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.BranchTODOs) == 0 {
		t.Error("expected BranchTODOs from base diff")
	}
	if b.BranchTODOs[0].Kind != "TODO" {
		t.Errorf("Kind = %q, want TODO", b.BranchTODOs[0].Kind)
	}
}

func TestCollect_WithBase_Empty(t *testing.T) {
	dir := tempGitRepo(t)
	base, err := currentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}

	b, err := Collect(Config{Dir: dir, NumCommits: 5, Base: base})
	if err != nil {
		t.Fatal(err)
	}
	if b.BaseBranch != base {
		t.Errorf("BaseBranch = %q, want %q", b.BaseBranch, base)
	}
	if len(b.BranchCommits) != 0 {
		t.Errorf("expected 0 branch commits, got %d", len(b.BranchCommits))
	}
	if b.BranchDiff != "" {
		t.Errorf("expected empty BranchDiff, got %q", b.BranchDiff)
	}
}

func TestCollect_WithTree(t *testing.T) {
	dir := tempGitRepo(t)
	b, err := Collect(Config{Dir: dir, NumCommits: 5, ShowTree: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Tree) == 0 {
		t.Error("expected Tree to be populated when ShowTree=true")
	}
	found := false
	for _, line := range b.Tree {
		if strings.Contains(line, "main.go") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected main.go in tree, got %v", b.Tree)
	}
}

func TestCollect_WithTree_NotSet(t *testing.T) {
	dir := tempGitRepo(t)
	b, err := Collect(Config{Dir: dir, NumCommits: 5, ShowTree: false})
	if err != nil {
		t.Fatal(err)
	}
	if len(b.Tree) != 0 {
		t.Errorf("expected empty Tree when ShowTree=false, got %v", b.Tree)
	}
}

func TestCollect_TokenBudget_TruncatesDiff(t *testing.T) {
	dir := tempGitRepo(t)
	base, err := currentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}

	mustGit(t, dir, "checkout", "-b", "feature")
	content := "package main\n" + strings.Repeat("// line of code\n", 1000)
	if err := os.WriteFile(filepath.Join(dir, "big.go"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "big commit")

	b, err := Collect(Config{Dir: dir, NumCommits: 5, Base: base, MaxTokens: 300})
	if err != nil {
		t.Fatal(err)
	}
	if estimateTokens(b.BranchDiff) > 300 {
		t.Errorf("BranchDiff too large after budget: %d tokens", estimateTokens(b.BranchDiff))
	}
}

func TestPrintText_WithBase(t *testing.T) {
	b := &Brief{
		Dir:        "/repo",
		Branch:     "feature",
		Commits:    []Commit{{Hash: "abc1234", Subject: "recent work"}},
		BaseBranch: "main",
		BranchCommits: []Commit{
			{Hash: "def5678", Subject: "branch commit one"},
		},
		BranchDiff:  "+package main\n+// new content here",
		BranchTODOs: []TODOItem{{File: "new.go", Line: 5, Kind: "TODO", Text: "finish this"}},
	}

	var buf bytes.Buffer
	if err := printText(&buf, b); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	for _, want := range []string{
		"branch commits vs main",
		"branch commit one",
		"diff vs main",
		"new content here",
		"todos in branch",
		"new.go:5",
		"TODO",
		"finish this",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintText_NoBase_NoBaseSections(t *testing.T) {
	b := &Brief{
		Dir:     "/repo",
		Branch:  "main",
		Commits: []Commit{{Hash: "abc1234", Subject: "a commit"}},
	}

	var buf bytes.Buffer
	if err := printText(&buf, b); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	for _, absent := range []string{"branch commits vs", "diff vs", "todos in branch"} {
		if strings.Contains(out, absent) {
			t.Errorf("output should not contain %q when no base set:\n%s", absent, out)
		}
	}
}

func TestPrintText_WithTree(t *testing.T) {
	b := &Brief{
		Dir:    "/repo",
		Branch: "main",
		Tree:   []string{"brief.go", "cmd/", "  brief/", "    main.go"},
	}

	var buf bytes.Buffer
	if err := printText(&buf, b); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	for _, want := range []string{"repo tree", "brief.go", "cmd/"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrintText_NoTree_NoTreeSection(t *testing.T) {
	b := &Brief{Dir: "/repo", Branch: "main"}

	var buf bytes.Buffer
	if err := printText(&buf, b); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "repo tree") {
		t.Errorf("output should not contain repo tree section:\n%s", out)
	}
}

func TestRun_WithBase(t *testing.T) {
	dir := tempGitRepo(t)
	base, err := currentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := Run([]string{"--dir", dir, "--base", base}); err != nil {
		t.Fatalf("Run with --base: %v", err)
	}
}

func TestRun_WithBase_BadRef(t *testing.T) {
	dir := tempGitRepo(t)
	// nonexistent base should not crash - graceful skip
	if err := Run([]string{"--dir", dir, "--base", "nonexistent-branch-xyz"}); err != nil {
		t.Fatalf("Run with bad --base should not error: %v", err)
	}
}

func TestRun_WithTree(t *testing.T) {
	dir := tempGitRepo(t)
	if err := Run([]string{"--dir", dir, "--tree"}); err != nil {
		t.Fatalf("Run with --tree: %v", err)
	}
}

func TestRun_JSON(t *testing.T) {
	dir := tempGitRepo(t)
	if err := Run([]string{"--dir", dir, "--json"}); err != nil {
		t.Fatalf("Run with --json: %v", err)
	}
}

func TestRun_Help(t *testing.T) {
	// -h should return an error (flag.ErrHelp) but not panic
	if err := Run([]string{"-h"}); err == nil {
		t.Error("expected error for -h")
	}
}

func TestRun_InvalidDir(t *testing.T) {
	if err := Run([]string{"--dir", "/nonexistent/path/xyz"}); err == nil {
		t.Error("expected error for invalid directory")
	}
}
