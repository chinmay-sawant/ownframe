package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSampleDiffShape(t *testing.T) {
	t.Parallel()

	cs := sampleChangeSet()

	if len(cs.Files) != 7 {
		t.Fatalf("files = %d, want 7", len(cs.Files))
	}

	if !cs.Sample {
		t.Fatal("sample flag lost")
	}

	add, del := cs.totals()
	if add != 562 || del != 109 {
		t.Fatalf("totals = +%d -%d, want +562 -109", add, del)
	}

	want := map[string]string{
		"backend/probe.go":     "M",
		"backend/store.go":     "A",
		"config/services.yaml": "M",
		"docs/deployment.md":   "M",
		"tools/legacy.py":      "D",
		"tools/report.py":      "M",
		"web/theme.css":        "A",
	}

	for _, file := range cs.Files {
		if got, ok := want[file.Path]; !ok {
			t.Errorf("unexpected file %s", file.Path)
		} else if file.Status != got {
			t.Errorf("%s status = %s, want %s", file.Path, file.Status, got)
		}
	}

	deployment := findFile(cs.Files, "docs/deployment.md")
	if deployment == nil || len(deployment.Hunks) != 2 {
		t.Fatalf("deployment hunks = %v", deployment)
	}

	if deployment.Add != 77 || deployment.Del != 11 {
		t.Fatalf("deployment stats = +%d -%d", deployment.Add, deployment.Del)
	}
}

func TestHunkNumbersAndText(t *testing.T) {
	t.Parallel()

	raw := `diff --git a/x.go b/x.go
index 111..222 100644
--- a/x.go
+++ b/x.go
@@ -2,2 +2,3 @@ func run() {
-old
+new
+more
 keep
`
	files := parseDiff(raw)
	if len(files) != 1 {
		t.Fatalf("files = %d", len(files))
	}

	h := files[0].Hunks[0]
	if h.OldStart != 2 || h.OldCount != 2 || h.NewStart != 2 || h.NewCount != 3 {
		t.Fatalf("hunk range = %+v", h)
	}

	if h.Header != "func run() {" {
		t.Fatalf("header = %q", h.Header)
	}

	got := []string{}
	for _, line := range h.Lines {
		got = append(got, string(line.Kind)+line.Text)
	}

	want := []string{"-old", "+new", "+more", " keep"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("lines = %v, want %v", got, want)
	}

	if h.Lines[0].OldNo != 2 || h.Lines[1].NewNo != 2 || h.Lines[3].OldNo != 3 || h.Lines[3].NewNo != 4 {
		t.Fatalf("numbers = %+v", h.Lines)
	}
}

func TestWordEmphasis(t *testing.T) {
	t.Parallel()

	raw := `diff --git a/x.go b/x.go
--- a/x.go
+++ b/x.go
@@ -1,1 +1,1 @@
-const total = 10
+const total = 20
`
	file := parseDiff(raw)[0]
	if len(file.Hunks) != 1 || len(file.Hunks[0].Lines) != 2 {
		t.Fatalf("hunk shape = %+v", file.Hunks)
	}

	del := file.Hunks[0].Lines[0]
	add := file.Hunks[0].Lines[1]
	if !del.Emph || !add.Emph {
		t.Fatalf("emphasis = del %v add %v", del.Emph, add.Emph)
	}

	delSegs := diffSegments(del, langGo)
	addSegs := diffSegments(add, langGo)

	if !hasSegment(delSegs, "10", "em-d") {
		t.Fatalf("deleted range not marked: %+v", delSegs)
	}

	if !hasSegment(addSegs, "20", "em-a") {
		t.Fatalf("added range not marked: %+v", addSegs)
	}
}

func hasSegment(segs []segment, text, class string) bool {
	for _, seg := range segs {
		if strings.TrimSpace(seg.Text) == text && strings.Contains(seg.Class, class) {
			return true
		}
	}

	return false
}

func TestSplitPairing(t *testing.T) {
	t.Parallel()

	raw := `diff --git a/x.go b/x.go
--- a/x.go
+++ b/x.go
@@ -1,2 +1,2 @@
-old one
-old two
+new one
+new two
 same
`
	file := parseDiff(raw)[0]
	rows := diffRowsSplit(file, langGo)

	var split []diffRow

	for _, row := range rows {
		if row.Kind == "split" {
			split = append(split, row)
		}
	}

	if len(split) != 3 {
		t.Fatalf("split rows = %d, want 3", len(split))
	}

	if split[0].LeftKind != rowDel || split[0].RightKind != rowAdd {
		t.Fatalf("row 0 kinds = %s/%s", split[0].LeftKind, split[0].RightKind)
	}

	if split[2].LeftKind != rowCtx || split[2].RightKind != rowCtx {
		t.Fatalf("context row kinds = %s/%s", split[2].LeftKind, split[2].RightKind)
	}

	if split[0].LeftNo != "1" || split[0].RightNo != "1" {
		t.Fatalf("row 0 numbers = %s/%s", split[0].LeftNo, split[0].RightNo)
	}
}

func TestExpandTabs(t *testing.T) {
	t.Parallel()

	if got := expandTabs("\tfunc\tmain"); got != "    func    main" {
		t.Fatalf("expandTabs = %q", got)
	}

	if got := expandTabs("plain\r"); got != "plain" {
		t.Fatalf("carriage return = %q", got)
	}
}

func TestUntrackedChange(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "notes.txt")

	if err := os.WriteFile(path, []byte("first\nsecond\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	file, ok := untrackedChange(dir, "notes.txt")
	if !ok {
		t.Fatal("untracked file skipped")
	}

	if file.Status != "U" || file.Add != 2 || file.Del != 0 {
		t.Fatalf("untracked = %+v", file)
	}

	if len(file.Hunks) != 1 || len(file.Hunks[0].Lines) != 2 {
		t.Fatalf("hunk = %+v", file.Hunks)
	}

	if file.Hunks[0].Lines[1].NewNo != 2 {
		t.Fatalf("line number = %d", file.Hunks[0].Lines[1].NewNo)
	}
}

func TestHunkLabel(t *testing.T) {
	t.Parallel()

	cases := []struct {
		h    hunk
		want string
	}{
		{hunk{OldStart: 1, OldCount: 1, NewStart: 1, NewCount: 1}, "@@ -1 +1 @@"},
		{hunk{OldStart: 10, OldCount: 26, NewStart: 10, NewCount: 92, Header: "cd backend"}, "@@ -10,26 +10,92 @@ cd backend"},
		{hunk{OldStart: 0, OldCount: 0, NewStart: 1, NewCount: 183}, "@@ -0,0 +1,183 @@"},
	}

	for _, tc := range cases {
		if got := hunkLabel(tc.h); got != tc.want {
			t.Errorf("hunkLabel = %q, want %q", got, tc.want)
		}
	}
}
