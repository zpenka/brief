package brief

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Config holds the options passed to Run.
type Config struct {
	Dir        string
	NumCommits int
	RunTests   bool
	JSON       bool
}

// Brief is the collected snapshot of a repository's current state.
type Brief struct {
	Dir         string       `json:"dir"`
	Branch      string       `json:"branch"`
	Commits     []Commit     `json:"commits"`
	Status      []StatusLine `json:"status"`
	TODOs       []TODOItem   `json:"todos"`
	StagedTODOs []TODOItem   `json:"staged_todos,omitempty"`
	Tests       *TestResult  `json:"tests,omitempty"`
	At          time.Time    `json:"at"`
}

// Run parses args and runs the tool, writing output to stdout.
func Run(args []string) error {
	fs := flag.NewFlagSet("brief", flag.ContinueOnError)
	dir := fs.String("dir", ".", "git repository directory")
	n := fs.Int("commits", 5, "number of recent commits to show")
	tests := fs.Bool("tests", false, "run tests and include result (slow)")
	jsonOut := fs.Bool("json", false, "output JSON instead of text")

	if err := fs.Parse(args); err != nil {
		return err
	}

	cfg := Config{
		Dir:        *dir,
		NumCommits: *n,
		RunTests:   *tests,
		JSON:       *jsonOut,
	}

	b, err := Collect(cfg)
	if err != nil {
		return err
	}

	if cfg.JSON {
		return printJSON(os.Stdout, b)
	}
	return printText(os.Stdout, b)
}

// Collect gathers all sections and returns a Brief. It is safe to call
// directly from other tools.
func Collect(cfg Config) (*Brief, error) {
	absDir, err := filepath.Abs(cfg.Dir)
	if err != nil {
		absDir = cfg.Dir
	}

	b := &Brief{
		Dir: absDir,
		At:  time.Now(),
	}

	b.Branch, err = currentBranch(absDir)
	if err != nil {
		return nil, fmt.Errorf("git branch: %w (is %q a git repository?)", err, absDir)
	}

	b.Commits, err = recentCommits(absDir, cfg.NumCommits)
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	b.Status, err = workingStatus(absDir)
	if err != nil {
		return nil, fmt.Errorf("git status: %w", err)
	}

	if diff, err := diffOutput(absDir); err == nil {
		b.TODOs = todosInDiff(diff)
	}

	if staged, err := stagedDiffOutput(absDir); err == nil {
		b.StagedTODOs = todosInDiff(staged)
	}

	if cfg.RunTests {
		b.Tests, _ = runTests(absDir, 60*time.Second)
	}

	return b, nil
}

func printText(w io.Writer, b *Brief) error {
	fmt.Fprintf(w, "# brief: %s\n\n", b.Dir)
	fmt.Fprintf(w, "**branch:** %s\n", b.Branch)
	fmt.Fprintf(w, "**as of:** %s\n", b.At.Format("2006-01-02 15:04"))

	fmt.Fprintf(w, "\n## recent commits\n")
	if len(b.Commits) == 0 {
		fmt.Fprintf(w, "none\n")
	}
	for _, c := range b.Commits {
		fmt.Fprintf(w, "- `%s` %s\n", c.Hash, c.Subject)
	}

	fmt.Fprintf(w, "\n## working tree\n")
	if len(b.Status) == 0 {
		fmt.Fprintf(w, "clean\n")
	}
	for _, s := range b.Status {
		fmt.Fprintf(w, "- %s: `%s`\n", s.Label(), s.Path)
	}

	if len(b.TODOs) > 0 {
		fmt.Fprintf(w, "\n## todos in recent changes\n")
		for _, td := range b.TODOs {
			fmt.Fprintf(w, "- `%s:%d` **%s**: %s\n", td.File, td.Line, td.Kind, td.Text)
		}
	}

	if len(b.StagedTODOs) > 0 {
		fmt.Fprintf(w, "\n## todos in staged changes\n")
		for _, td := range b.StagedTODOs {
			fmt.Fprintf(w, "- `%s:%d` **%s**: %s\n", td.File, td.Line, td.Kind, td.Text)
		}
	}

	if b.Tests != nil {
		fmt.Fprintf(w, "\n## tests\n")
		if b.Tests.Passed {
			fmt.Fprintf(w, "✓ %d passed (%s)\n", b.Tests.Count, b.Tests.Elapsed.Round(time.Millisecond))
		} else {
			fmt.Fprintf(w, "✗ failed\n")
			detailsByTest := make(map[string][]string, len(b.Tests.Details))
			for _, d := range b.Tests.Details {
				detailsByTest[d.Test] = d.Messages
			}
			for _, f := range b.Tests.Failed {
				fmt.Fprintf(w, "- %s\n", f)
				for _, msg := range detailsByTest[f] {
					fmt.Fprintf(w, "  %s", msg)
				}
			}
		}
	}

	return nil
}

func printJSON(w io.Writer, b *Brief) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(b)
}
