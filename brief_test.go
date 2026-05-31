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
