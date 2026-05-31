package brief

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

// FailureDetail holds the output messages for a single failing test.
type FailureDetail struct {
	Test     string   `json:"test"`
	Messages []string `json:"messages"`
}

// TestResult holds the outcome of running go test.
type TestResult struct {
	Passed  bool            `json:"passed"`
	Count   int             `json:"count"`
	Elapsed time.Duration   `json:"elapsed_ns"`
	Failed  []string        `json:"failed,omitempty"`
	Details []FailureDetail `json:"details,omitempty"`
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
	Output  string  `json:"Output"`
}

func parseTestEvents(output string) *TestResult {
	var (
		r       TestResult
		elapsed float64
		hasData bool
		// accumulate output lines per test name
		testOutput = map[string][]string{}
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
		case "output":
			if ev.Test != "" && ev.Output != "" {
				testOutput[ev.Test] = append(testOutput[ev.Test], ev.Output)
			}
		case "pass":
			if ev.Test != "" {
				r.Count++
				delete(testOutput, ev.Test)
			} else {
				elapsed += ev.Elapsed
			}
		case "fail":
			if ev.Test != "" {
				r.Failed = append(r.Failed, ev.Test)
				if msgs := testOutput[ev.Test]; len(msgs) > 0 {
					r.Details = append(r.Details, FailureDetail{Test: ev.Test, Messages: msgs})
				}
				delete(testOutput, ev.Test)
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
