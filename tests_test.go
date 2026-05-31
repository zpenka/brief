package brief

import (
	"strings"
	"testing"
	"time"
)

func TestParseTestEvents_AllPass(t *testing.T) {
	input := `{"Action":"start","Package":"github.com/zpenka/brief"}
{"Action":"run","Package":"github.com/zpenka/brief","Test":"TestFoo"}
{"Action":"pass","Package":"github.com/zpenka/brief","Test":"TestFoo","Elapsed":0.001}
{"Action":"run","Package":"github.com/zpenka/brief","Test":"TestBar"}
{"Action":"pass","Package":"github.com/zpenka/brief","Test":"TestBar","Elapsed":0.002}
{"Action":"pass","Package":"github.com/zpenka/brief","Elapsed":0.005}
`
	r := parseTestEvents(input)
	if r == nil {
		t.Fatal("nil result")
	}
	if !r.Passed {
		t.Error("want Passed = true")
	}
	if r.Count != 2 {
		t.Errorf("Count = %d, want 2", r.Count)
	}
	if len(r.Failed) != 0 {
		t.Errorf("Failed = %v, want empty", r.Failed)
	}
	if r.Elapsed < 4*time.Millisecond || r.Elapsed > 10*time.Millisecond {
		t.Errorf("Elapsed = %v, expected ~5ms", r.Elapsed)
	}
}

func TestParseTestEvents_WithFailure(t *testing.T) {
	input := `{"Action":"run","Package":"github.com/zpenka/brief","Test":"TestFoo"}
{"Action":"fail","Package":"github.com/zpenka/brief","Test":"TestFoo","Elapsed":0.001}
{"Action":"run","Package":"github.com/zpenka/brief","Test":"TestBar"}
{"Action":"pass","Package":"github.com/zpenka/brief","Test":"TestBar","Elapsed":0.002}
{"Action":"fail","Package":"github.com/zpenka/brief","Elapsed":0.005}
`
	r := parseTestEvents(input)
	if r == nil {
		t.Fatal("nil result")
	}
	if r.Passed {
		t.Error("want Passed = false")
	}
	if r.Count != 1 {
		t.Errorf("Count = %d, want 1", r.Count)
	}
	if len(r.Failed) != 1 || r.Failed[0] != "TestFoo" {
		t.Errorf("Failed = %v, want [TestFoo]", r.Failed)
	}
}

func TestParseTestEvents_FailureDetails(t *testing.T) {
	input := `{"Action":"run","Package":"p","Test":"TestFoo"}
{"Action":"output","Package":"p","Test":"TestFoo","Output":"    foo_test.go:42: want true, got false\n"}
{"Action":"output","Package":"p","Test":"TestFoo","Output":"    foo_test.go:43: another error\n"}
{"Action":"fail","Package":"p","Test":"TestFoo","Elapsed":0.001}
{"Action":"fail","Package":"p","Elapsed":0.005}
`
	r := parseTestEvents(input)
	if r == nil {
		t.Fatal("nil result")
	}
	if r.Passed {
		t.Error("want Passed = false")
	}
	if len(r.Details) != 1 {
		t.Fatalf("len(Details) = %d, want 1", len(r.Details))
	}
	if r.Details[0].Test != "TestFoo" {
		t.Errorf("Details[0].Test = %q, want TestFoo", r.Details[0].Test)
	}
	if len(r.Details[0].Messages) != 2 {
		t.Fatalf("len(Messages) = %d, want 2", len(r.Details[0].Messages))
	}
	if !strings.Contains(r.Details[0].Messages[0], "want true, got false") {
		t.Errorf("Messages[0] = %q", r.Details[0].Messages[0])
	}
}

func TestParseTestEvents_FailureDetails_MultipleTests(t *testing.T) {
	input := `{"Action":"run","Package":"p","Test":"TestA"}
{"Action":"output","Package":"p","Test":"TestA","Output":"    a_test.go:1: err A\n"}
{"Action":"fail","Package":"p","Test":"TestA","Elapsed":0.001}
{"Action":"run","Package":"p","Test":"TestB"}
{"Action":"output","Package":"p","Test":"TestB","Output":"    b_test.go:2: err B\n"}
{"Action":"fail","Package":"p","Test":"TestB","Elapsed":0.002}
{"Action":"run","Package":"p","Test":"TestC"}
{"Action":"pass","Package":"p","Test":"TestC","Elapsed":0.001}
{"Action":"fail","Package":"p","Elapsed":0.005}
`
	r := parseTestEvents(input)
	if r == nil {
		t.Fatal("nil result")
	}
	if len(r.Details) != 2 {
		t.Fatalf("len(Details) = %d, want 2", len(r.Details))
	}
	if r.Details[0].Test != "TestA" || r.Details[1].Test != "TestB" {
		t.Errorf("Details tests = %q, %q", r.Details[0].Test, r.Details[1].Test)
	}
}

func TestParseTestEvents_PassedTestsHaveNoDetails(t *testing.T) {
	input := `{"Action":"run","Package":"p","Test":"TestFoo"}
{"Action":"output","Package":"p","Test":"TestFoo","Output":"=== RUN TestFoo\n"}
{"Action":"pass","Package":"p","Test":"TestFoo","Elapsed":0.001}
{"Action":"pass","Package":"p","Elapsed":0.005}
`
	r := parseTestEvents(input)
	if r == nil {
		t.Fatal("nil result")
	}
	if len(r.Details) != 0 {
		t.Errorf("expected no Details for passing tests, got %v", r.Details)
	}
}

func TestParseTestEvents_Empty(t *testing.T) {
	if r := parseTestEvents(""); r != nil {
		t.Errorf("want nil for empty input, got %+v", r)
	}
}

func TestParseTestEvents_BuildFailure(t *testing.T) {
	// build failures produce output action with no test-level pass/fail
	input := `{"Action":"build-fail","Package":"github.com/zpenka/brief"}
{"Action":"fail","Package":"github.com/zpenka/brief","Elapsed":0.001}
`
	r := parseTestEvents(input)
	if r == nil {
		t.Fatal("nil result")
	}
	if r.Passed {
		t.Error("want Passed = false for build failure")
	}
}
