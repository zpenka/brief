package brief

import (
	"bytes"
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
