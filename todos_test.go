package brief

import "testing"

func TestTodosInDiff_Empty(t *testing.T) {
	if got := todosInDiff(""); len(got) != 0 {
		t.Errorf("want empty, got %v", got)
	}
}

func TestTodosInDiff_NoMarkers(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
index abc..def 100644
--- a/main.go
+++ b/main.go
@@ -1,2 +1,3 @@
 package main
+// just a comment
 func main() {}
`
	if got := todosInDiff(diff); len(got) != 0 {
		t.Errorf("want empty, got %v", got)
	}
}

func TestTodosInDiff_FindsTODO(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
index abc..def 100644
--- a/main.go
+++ b/main.go
@@ -1,2 +1,3 @@
 package main
+// TODO: implement this properly
 func main() {}
`
	got := todosInDiff(diff)
	if len(got) != 1 {
		t.Fatalf("len = %d, want 1", len(got))
	}
	if got[0].File != "main.go" {
		t.Errorf("File = %q, want %q", got[0].File, "main.go")
	}
	if got[0].Kind != "TODO" {
		t.Errorf("Kind = %q, want TODO", got[0].Kind)
	}
	if got[0].Text != "implement this properly" {
		t.Errorf("Text = %q", got[0].Text)
	}
	if got[0].Line != 2 {
		t.Errorf("Line = %d, want 2", got[0].Line)
	}
}

func TestTodosInDiff_MultipleMarkers(t *testing.T) {
	diff := `diff --git a/auth.go b/auth.go
index abc..def 100644
--- a/auth.go
+++ b/auth.go
@@ -10,3 +10,5 @@
 func validate() {
+	// FIXME: handle expired tokens
+	// HACK: temporary workaround
 	return nil
 }
`
	got := todosInDiff(diff)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].Kind != "FIXME" || got[0].Line != 11 {
		t.Errorf("got[0] = %+v", got[0])
	}
	if got[1].Kind != "HACK" || got[1].Line != 12 {
		t.Errorf("got[1] = %+v", got[1])
	}
}

func TestTodosInDiff_MultipleFiles(t *testing.T) {
	diff := `diff --git a/a.go b/a.go
index abc..def 100644
--- a/a.go
+++ b/a.go
@@ -1,1 +1,2 @@
 package main
+// TODO: file a
diff --git a/b.go b/b.go
index abc..def 100644
--- a/b.go
+++ b/b.go
@@ -1,1 +1,2 @@
 package main
+// NOTE: file b
`
	got := todosInDiff(diff)
	if len(got) != 2 {
		t.Fatalf("len = %d, want 2", len(got))
	}
	if got[0].File != "a.go" {
		t.Errorf("got[0].File = %q", got[0].File)
	}
	if got[1].File != "b.go" {
		t.Errorf("got[1].File = %q", got[1].File)
	}
}

func TestParseHunkStart(t *testing.T) {
	tests := []struct {
		line string
		want int
	}{
		{"@@ -1,3 +1,4 @@ func foo()", 1},
		{"@@ -0,0 +1 @@", 1},
		{"@@ -10,5 +12,7 @@", 12},
		{"@@ -1,2 +42,3 @@ some context", 42},
	}
	for _, tt := range tests {
		got := parseHunkStart(tt.line)
		if got != tt.want {
			t.Errorf("parseHunkStart(%q) = %d, want %d", tt.line, got, tt.want)
		}
	}
}

func TestFindTODO(t *testing.T) {
	tests := []struct {
		line     string
		wantKind string
		wantText string
		wantOK   bool
	}{
		{"// TODO: implement", "TODO", "implement", true},
		{"// FIXME: broken", "FIXME", "broken", true},
		{"// HACK: workaround", "HACK", "workaround", true},
		{"// NOTE: see docs", "NOTE", "see docs", true},
		{"// XXX: remove later", "XXX", "remove later", true},
		{"// just a comment", "", "", false},
		{"return nil", "", "", false},
		{"// todo: lowercase", "TODO", "lowercase", true},
	}
	for _, tt := range tests {
		kind, text, ok := findTODO(tt.line)
		if ok != tt.wantOK {
			t.Errorf("findTODO(%q): ok = %v, want %v", tt.line, ok, tt.wantOK)
			continue
		}
		if ok && kind != tt.wantKind {
			t.Errorf("findTODO(%q): kind = %q, want %q", tt.line, kind, tt.wantKind)
		}
		if ok && text != tt.wantText {
			t.Errorf("findTODO(%q): text = %q, want %q", tt.line, text, tt.wantText)
		}
	}
}
