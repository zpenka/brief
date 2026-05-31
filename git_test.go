package brief

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseBranch(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"main\n", "main"},
		{"feature/my-branch\n", "feature/my-branch"},
		{"HEAD\n", "HEAD"},
		{"  main  \n", "main"},
		{"", ""},
	}
	for _, tt := range tests {
		got := parseBranch(tt.input)
		if got != tt.want {
			t.Errorf("parseBranch(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseCommits(t *testing.T) {
	input := "abc1234 fix auth validation\ndef5678 add user endpoints\n"
	got := parseCommits(input)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Hash != "abc1234" || got[0].Subject != "fix auth validation" {
		t.Errorf("got[0] = %+v", got[0])
	}
	if got[1].Hash != "def5678" || got[1].Subject != "add user endpoints" {
		t.Errorf("got[1] = %+v", got[1])
	}
}

func TestParseCommitsEmpty(t *testing.T) {
	if got := parseCommits(""); len(got) != 0 {
		t.Errorf("want empty, got %v", got)
	}
}

func TestParseCommitsMultiWordSubject(t *testing.T) {
	input := "abc1234 fix: handle edge case in auth -> redirect flow\n"
	got := parseCommits(input)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].Subject != "fix: handle edge case in auth -> redirect flow" {
		t.Errorf("Subject = %q", got[0].Subject)
	}
}

func TestParseStatus(t *testing.T) {
	input := " M src/auth.go\n?? new_file.go\nA  staged.go\n"
	got := parseStatus(input)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	if got[0].XY != " M" || got[0].Path != "src/auth.go" {
		t.Errorf("got[0] = %+v", got[0])
	}
	if got[1].XY != "??" || got[1].Path != "new_file.go" {
		t.Errorf("got[1] = %+v", got[1])
	}
	if got[2].XY != "A " || got[2].Path != "staged.go" {
		t.Errorf("got[2] = %+v", got[2])
	}
}

func TestParseStatusEmpty(t *testing.T) {
	if got := parseStatus(""); len(got) != 0 {
		t.Errorf("want empty, got %v", got)
	}
}

func TestStatusLineLabel(t *testing.T) {
	tests := []struct {
		xy   string
		want string
	}{
		{"??", "untracked"},
		{" M", "modified"},
		{"M ", "modified"},
		{"MM", "modified"},
		{"A ", "added"},
		{" D", "deleted"},
		{"D ", "deleted"},
		{"R ", "renamed"},
		{"UU", "conflict"},
		{"AU", "conflict"},
	}
	for _, tt := range tests {
		s := StatusLine{XY: tt.xy}
		if got := s.Label(); got != tt.want {
			t.Errorf("Label(%q) = %q, want %q", tt.xy, got, tt.want)
		}
	}
}

// tempGitRepo creates an initialized git repo with one commit in a temp dir.
func tempGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, args := range [][]string{
		{"init"},
		{"config", "user.email", "test@test.com"},
		{"config", "user.name", "Test"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	// initial commit
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n"), 0644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{
		{"add", "."},
		{"commit", "-m", "initial commit"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return dir
}

func TestCurrentBranch(t *testing.T) {
	dir := tempGitRepo(t)
	branch, err := currentBranch(dir)
	if err != nil {
		t.Fatal(err)
	}
	if branch == "" {
		t.Error("expected non-empty branch")
	}
}

func TestRecentCommits(t *testing.T) {
	dir := tempGitRepo(t)
	commits, err := recentCommits(dir, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) == 0 {
		t.Error("expected at least one commit")
	}
	if commits[0].Hash == "" || commits[0].Subject == "" {
		t.Errorf("empty commit fields: %+v", commits[0])
	}
}

func TestWorkingStatus(t *testing.T) {
	dir := tempGitRepo(t)

	// clean tree
	status, err := workingStatus(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(status) != 0 {
		t.Errorf("expected clean status, got %v", status)
	}

	// make a change
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n// changed\n"), 0644); err != nil {
		t.Fatal(err)
	}
	status, err = workingStatus(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(status) == 0 {
		t.Error("expected dirty status after modification")
	}
}
