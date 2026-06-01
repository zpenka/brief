package brief

// estimateTokens returns a rough token count for s using the ~4 chars/token heuristic.
func estimateTokens(s string) int {
	return len(s) / 4
}

func estimateBriefTokens(b *Brief) int {
	total := 50 // base overhead for headers and formatting
	total += estimateTokens(b.Branch) + estimateTokens(b.Dir)
	for _, c := range b.Commits {
		total += estimateTokens(c.Hash + " " + c.Subject)
	}
	for _, c := range b.BranchCommits {
		total += estimateTokens(c.Hash + " " + c.Subject)
	}
	for _, s := range b.Status {
		total += estimateTokens(s.Path)
	}
	total += estimateTokens(b.BranchDiff)
	for _, td := range b.TODOs {
		total += estimateTokens(td.File + td.Text)
	}
	for _, td := range b.StagedTODOs {
		total += estimateTokens(td.File + td.Text)
	}
	for _, td := range b.BranchTODOs {
		total += estimateTokens(td.File + td.Text)
	}
	for _, line := range b.Tree {
		total += estimateTokens(line)
	}
	return total
}

// applyBudget truncates b in place to fit within maxTokens.
// It trims BranchDiff first (largest), then Commits, then BranchCommits.
// maxTokens <= 0 is a no-op.
func applyBudget(b *Brief, maxTokens int) {
	if maxTokens <= 0 {
		return
	}

	// Truncate BranchDiff first — it's the most variable-length field.
	const truncMarker = "\n... [truncated]"
	if len(b.BranchDiff) > 0 {
		nonDiff := estimateBriefTokens(b) - estimateTokens(b.BranchDiff)
		available := maxTokens - nonDiff
		if available < 0 {
			available = 0
		}
		// Reserve chars for the marker so the total stays within budget.
		maxChars := available*4 - len(truncMarker)
		if maxChars < 0 {
			maxChars = 0
		}
		if len(b.BranchDiff) > maxChars {
			if maxChars > 0 {
				b.BranchDiff = b.BranchDiff[:maxChars] + truncMarker
			} else {
				b.BranchDiff = "... [truncated]"
			}
		}
	}

	// Trim commits until under budget (keep at least one).
	for estimateBriefTokens(b) > maxTokens && len(b.Commits) > 1 {
		b.Commits = b.Commits[:len(b.Commits)-1]
	}

	// Trim branch commits until under budget (keep at least one).
	for estimateBriefTokens(b) > maxTokens && len(b.BranchCommits) > 1 {
		b.BranchCommits = b.BranchCommits[:len(b.BranchCommits)-1]
	}
}
