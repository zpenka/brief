package brief

import (
	"fmt"
	"os/exec"
	"strings"
)

// Commit is a single entry from git log --oneline.
type Commit struct {
	Hash    string `json:"hash"`
	Subject string `json:"subject"`
}

// StatusLine is a single entry from git status --porcelain.
type StatusLine struct {
	XY   string `json:"xy"`
	Path string `json:"path"`
}

// Label returns a human-readable description of the change type.
func (s StatusLine) Label() string {
	if len(s.XY) < 2 {
		return "modified"
	}
	x, y := s.XY[0], s.XY[1]
	if x == '?' && y == '?' {
		return "untracked"
	}
	if x == 'U' || y == 'U' || (x == 'A' && y == 'A') || (x == 'D' && y == 'D') {
		return "conflict"
	}
	if x == 'R' {
		return "renamed"
	}
	if x == 'D' || y == 'D' {
		return "deleted"
	}
	if x == 'A' {
		return "added"
	}
	return "modified"
}

func currentBranch(dir string) (string, error) {
	out, err := gitCmd(dir, "branch", "--show-current")
	if err != nil {
		return "", err
	}
	branch := parseBranch(out)
	if branch == "" {
		// detached HEAD
		hash, err := gitCmd(dir, "rev-parse", "--short", "HEAD")
		if err != nil {
			return "HEAD", nil
		}
		return "HEAD@" + strings.TrimSpace(hash), nil
	}
	return branch, nil
}

func recentCommits(dir string, n int) ([]Commit, error) {
	out, err := gitCmd(dir, "log", fmt.Sprintf("-n%d", n), "--oneline", "--no-decorate")
	if err != nil {
		return nil, err
	}
	return parseCommits(out), nil
}

func workingStatus(dir string) ([]StatusLine, error) {
	out, err := gitCmd(dir, "status", "--porcelain")
	if err != nil {
		return nil, err
	}
	return parseStatus(out), nil
}

func diffOutput(dir string) (string, error) {
	return gitCmd(dir, "diff", "HEAD")
}

func stagedDiffOutput(dir string) (string, error) {
	return gitCmd(dir, "diff", "--cached")
}

// commitsAhead returns commits reachable from HEAD but not from base.
func commitsAhead(dir, base string) ([]Commit, error) {
	out, err := gitCmd(dir, "log", base+"..HEAD", "--oneline", "--no-decorate")
	if err != nil {
		return nil, err
	}
	return parseCommits(out), nil
}

// baseDiff returns the diff of all changes introduced since branching from base.
func baseDiff(dir, base string) (string, error) {
	return gitCmd(dir, "diff", base+"...HEAD")
}

func gitCmd(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func parseBranch(output string) string {
	return strings.TrimSpace(output)
}

func parseCommits(output string) []Commit {
	var commits []Commit
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		hash, subject, ok := strings.Cut(line, " ")
		if !ok {
			continue
		}
		commits = append(commits, Commit{Hash: hash, Subject: subject})
	}
	return commits
}

func parseStatus(output string) []StatusLine {
	var lines []StatusLine
	for _, line := range strings.Split(output, "\n") {
		if len(line) < 4 {
			continue
		}
		lines = append(lines, StatusLine{
			XY:   line[:2],
			Path: line[3:],
		})
	}
	return lines
}
