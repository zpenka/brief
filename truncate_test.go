package brief

import (
	"strings"
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	cases := []struct {
		s    string
		want int
	}{
		{"", 0},
		{"aaaa", 1},
		{"aaa", 0},
		{strings.Repeat("x", 400), 100},
		{strings.Repeat("x", 401), 100},
		{strings.Repeat("x", 404), 101},
	}
	for _, tt := range cases {
		got := estimateTokens(tt.s)
		if got != tt.want {
			t.Errorf("estimateTokens(%d chars) = %d, want %d", len(tt.s), got, tt.want)
		}
	}
}

func TestApplyBudget_NoOp_WhenZero(t *testing.T) {
	b := &Brief{BranchDiff: strings.Repeat("x", 10000)}
	original := b.BranchDiff
	applyBudget(b, 0)
	if b.BranchDiff != original {
		t.Error("applyBudget with maxTokens=0 should not modify brief")
	}
}

func TestApplyBudget_NoOp_WhenUnderBudget(t *testing.T) {
	b := &Brief{
		Branch:     "main",
		BranchDiff: "small diff",
	}
	applyBudget(b, 10000)
	if b.BranchDiff != "small diff" {
		t.Errorf("small brief should not be modified, got %q", b.BranchDiff)
	}
}

func TestApplyBudget_TruncatesBranchDiff(t *testing.T) {
	b := &Brief{
		Branch:     "feature",
		BranchDiff: strings.Repeat("x", 40000),
	}
	applyBudget(b, 500)
	if estimateTokens(b.BranchDiff) > 500 {
		t.Errorf("BranchDiff not truncated: %d tokens", estimateTokens(b.BranchDiff))
	}
	if !strings.Contains(b.BranchDiff, "[truncated]") {
		t.Error("truncated diff should contain [truncated] marker")
	}
}

func TestApplyBudget_TruncatesCommits(t *testing.T) {
	commits := make([]Commit, 20)
	for i := range commits {
		commits[i] = Commit{Hash: "abc1234", Subject: strings.Repeat("long subject word ", 10)}
	}
	b := &Brief{
		Branch:  "main",
		Commits: commits,
	}
	applyBudget(b, 50)
	if len(b.Commits) >= 20 {
		t.Errorf("commits should be trimmed, got %d", len(b.Commits))
	}
}

func TestApplyBudget_KeepsAtLeastOneCommit(t *testing.T) {
	b := &Brief{
		Branch:  "main",
		Commits: []Commit{{Hash: "abc1234", Subject: "the only commit"}},
	}
	applyBudget(b, 1) // impossibly tight
	if len(b.Commits) == 0 {
		t.Error("should keep at least one commit")
	}
}

func TestApplyBudget_TruncatesBranchDiff_BeforeCommits(t *testing.T) {
	b := &Brief{
		Branch:     "feature",
		BranchDiff: strings.Repeat("x", 40000),
		Commits: []Commit{
			{Hash: "abc1234", Subject: "commit one"},
			{Hash: "def5678", Subject: "commit two"},
		},
	}
	applyBudget(b, 500)
	// Commits should be intact; diff should be truncated
	if len(b.Commits) != 2 {
		t.Errorf("commits should be intact when only diff needs trimming, got %d", len(b.Commits))
	}
	if !strings.Contains(b.BranchDiff, "[truncated]") {
		t.Error("diff should be truncated")
	}
}
