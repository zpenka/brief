package brief

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildTree_Empty(t *testing.T) {
	got := buildTree(nil, 3)
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestBuildTree_FlatFiles(t *testing.T) {
	paths := []string{"a.go", "b.go", "c.go"}
	got := buildTree(paths, 3)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	want := []string{"a.go", "b.go", "c.go"}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("[%d] got %q, want %q", i, got[i], w)
		}
	}
}

func TestBuildTree_Nested(t *testing.T) {
	paths := []string{
		"brief.go",
		"cmd/brief/main.go",
		"git.go",
	}
	got := buildTree(paths, 3)
	joined := strings.Join(got, "\n")

	for _, want := range []string{"cmd/", "  brief/", "    main.go"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in:\n%s", want, joined)
		}
	}
}

func TestBuildTree_MaxDepth(t *testing.T) {
	paths := []string{
		"a/b/c/deep.go",
		"top.go",
	}
	got := buildTree(paths, 2)
	joined := strings.Join(got, "\n")

	if strings.Contains(joined, "deep.go") {
		t.Errorf("file beyond max depth should be collapsed, got:\n%s", joined)
	}
	if !strings.Contains(joined, "a/") {
		t.Errorf("top-level dir should appear, got:\n%s", joined)
	}
}

func TestBuildTree_SortedAlphabetically(t *testing.T) {
	paths := []string{"z.go", "a.go", "m.go"}
	got := buildTree(paths, 3)
	if got[0] != "a.go" || got[1] != "m.go" || got[2] != "z.go" {
		t.Errorf("expected alphabetical order, got %v", got)
	}
}

func TestBuildTree_DirsBeforeFiles(t *testing.T) {
	paths := []string{"z.go", "a/file.go", "b.go"}
	got := buildTree(paths, 3)
	joined := strings.Join(got, "\n")
	aIdx := strings.Index(joined, "a/")
	bIdx := strings.Index(joined, "b.go")
	zIdx := strings.Index(joined, "z.go")
	if aIdx < 0 || bIdx < 0 || zIdx < 0 {
		t.Fatalf("missing entries: %s", joined)
	}
	if aIdx > bIdx {
		t.Errorf("dir a/ should appear before file b.go:\n%s", joined)
	}
	if aIdx > zIdx {
		t.Errorf("dir a/ should appear before file z.go:\n%s", joined)
	}
}

func TestRepoFiles_TrackedFile(t *testing.T) {
	dir := tempGitRepo(t)
	files, err := repoFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("expected at least one file")
	}
	found := false
	for _, f := range files {
		if f == "main.go" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected main.go in files, got %v", files)
	}
}

func TestRepoFiles_IncludesUntracked(t *testing.T) {
	dir := tempGitRepo(t)

	if err := os.WriteFile(filepath.Join(dir, "untracked.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}

	files, err := repoFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range files {
		if f == "untracked.go" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected untracked.go in files, got %v", files)
	}
}

func TestRepoFiles_Sorted(t *testing.T) {
	dir := tempGitRepo(t)

	for _, name := range []string{"z_extra.go", "a_extra.go"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("package main\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "add extras")

	files, err := repoFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i < len(files); i++ {
		if files[i-1] > files[i] {
			t.Errorf("files not sorted: %v", files)
		}
	}
}

func TestRepoFiles_SubdirFile(t *testing.T) {
	dir := tempGitRepo(t)

	if err := os.MkdirAll(filepath.Join(dir, "cmd", "app"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "cmd", "app", "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "add subdir")

	files, err := repoFiles(dir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range files {
		if f == "cmd/app/main.go" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected cmd/app/main.go in files, got %v", files)
	}
}
