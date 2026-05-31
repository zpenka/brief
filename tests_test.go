package brief

import (
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
