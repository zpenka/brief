package brief

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

// TestResult holds the outcome of running go test.
type TestResult struct {
	Passed  bool          `json:"passed"`
	Count   int           `json:"count"`
	Elapsed time.Duration `json:"elapsed_ns"`
	Failed  []string      `json:"failed,omitempty"`
}

func runTests(dir string, timeout time.Duration) (*TestResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "test", "-json", "-race", "./...")
	cmd.Dir = dir
	out, _ := cmd.Output() // exit code non-zero on failure; parse JSON instead

	r := parseTestEvents(string(out))
	if r == nil {
		return nil, nil
	}
	return r, nil
}

type testEvent struct {
	Action  string  `json:"Action"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
}

func parseTestEvents(output string) *TestResult {
	var (
		r       TestResult
		elapsed float64
		hasData bool
	)
	r.Passed = true

	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		var ev testEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		hasData = true
		switch ev.Action {
		case "pass":
			if ev.Test != "" {
				r.Count++
			} else {
				elapsed += ev.Elapsed
			}
		case "fail":
			if ev.Test != "" {
				r.Failed = append(r.Failed, ev.Test)
			} else {
				r.Passed = false
				elapsed += ev.Elapsed
			}
		}
	}

	if !hasData {
		return nil
	}
	r.Elapsed = time.Duration(elapsed * float64(time.Second))
	return &r
}
