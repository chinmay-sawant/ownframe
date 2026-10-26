package main

import "testing"

func TestInlineChangesAdded(t *testing.T) {
	t.Parallel()

	files := parseDiff(`diff --git a/x.go b/x.go
--- a/x.go
+++ b/x.go
@@ -1,2 +1,4 @@
 one
+two
+three
 four
`)
	if len(files) != 1 {
		t.Fatalf("files = %d, want 1", len(files))
	}

	changes := inlineChanges(files[0])
	if len(changes) != 2 {
		t.Fatalf("changes = %d, want 2: %+v", len(changes), changes)
	}

	checkInlineLine(t, changes, 2, inlineAdded, "")
	checkInlineLine(t, changes, 3, inlineAdded, "")

	if _, ok := inlineChangeAt(changes, 1); ok {
		t.Fatal("context line 1 was stored")
	}

	if _, ok := inlineChangeAt(changes, 4); ok {
		t.Fatal("context line 4 was stored")
	}
}

func TestInlineChangesModifiedPair(t *testing.T) {
	t.Parallel()

	files := parseDiff(`diff --git a/x.go b/x.go
--- a/x.go
+++ b/x.go
@@ -5,5 +5,5 @@
 keep
-old
+new
-goneA
-goneB
+goneC
+goneD
 tail
`)
	if len(files) != 1 {
		t.Fatalf("files = %d, want 1", len(files))
	}

	changes := inlineChanges(files[0])
	if len(changes) != 3 {
		t.Fatalf("changes = %d, want 3: %+v", len(changes), changes)
	}

	checkInlineLine(t, changes, 6, inlineModified, "old")
	checkInlineLine(t, changes, 7, inlineModified, "goneA")
	checkInlineLine(t, changes, 8, inlineModified, "goneB")

	if _, ok := inlineChangeAt(changes, 9); ok {
		t.Fatal("context line 9 was stored")
	}
}

func TestInlineChangesDeletedOnly(t *testing.T) {
	t.Parallel()

	files := parseDiff(`diff --git a/x.go b/x.go
--- a/x.go
+++ b/x.go
@@ -4,2 +4,1 @@
 keep
-removed
@@ -10,2 +10,2 @@
-dropped
 after
@@ -20,3 +20,0 @@
-a
-b
-c
`)
	if len(files) != 1 {
		t.Fatalf("files = %d, want 1", len(files))
	}

	changes := inlineChanges(files[0])
	if len(changes) != 3 {
		t.Fatalf("changes = %d, want 3: %+v", len(changes), changes)
	}

	checkInlineLine(t, changes, 4, inlineDeleted, "removed")
	checkInlineLine(t, changes, 10, inlineDeleted, "dropped")
	checkInlineLine(t, changes, 20, inlineDeleted, "a\nb\nc")

	if _, ok := inlineChangeAt(changes, 5); ok {
		t.Fatal("old line 5 was stored as a new line")
	}
}

func TestInlineChangesStats(t *testing.T) {
	t.Parallel()

	files := parseDiff(`diff --git a/x.go b/x.go
--- a/x.go
+++ b/x.go
@@ -1,1 +1,2 @@
 base
+added
@@ -10,1 +11,1 @@
-old
+new
@@ -20,3 +20,0 @@
-a
-b
-c
`)
	if len(files) != 1 {
		t.Fatalf("files = %d, want 1", len(files))
	}

	added, modified, deleted := inlineStats(inlineChanges(files[0]))
	if added != 1 || modified != 1 || deleted != 3 {
		t.Fatalf("stats = +%d ~%d -%d, want +1 ~1 -3", added, modified, deleted)
	}

	added, modified, deleted = inlineStats(nil)
	if added != 0 || modified != 0 || deleted != 0 {
		t.Fatalf("nil stats = +%d ~%d -%d", added, modified, deleted)
	}

	added, modified, deleted = inlineStats(map[int]inlineLine{
		1: {Kind: inlineNone},
		2: {Kind: inlineAdded},
	})
	if added != 1 || modified != 0 || deleted != 0 {
		t.Fatalf("none stats = +%d ~%d -%d", added, modified, deleted)
	}
}

func TestInlineChangesBinaryAndEmpty(t *testing.T) {
	t.Parallel()

	files := parseDiff(`diff --git a/logo.png b/logo.png
Binary files a/logo.png and b/logo.png differ
`)
	if len(files) != 1 || !files[0].Binary {
		t.Fatalf("binary file = %+v", files)
	}

	if got := inlineChanges(files[0]); got != nil {
		t.Fatalf("binary changes = %+v, want nil", got)
	}

	files = parseDiff(`diff --git a/new.go b/new.go
new file mode 100644
--- /dev/null
+++ b/new.go
`)
	if len(files) != 1 || len(files[0].Hunks) != 0 {
		t.Fatalf("empty file = %+v", files)
	}

	if got := inlineChanges(files[0]); got != nil {
		t.Fatalf("empty changes = %+v, want nil", got)
	}

	binaryWithHunks := changedFile{
		Binary: true,
		Hunks: []hunk{{
			NewStart: 1,
			Lines:    []diffLine{{Kind: '+', NewNo: 1, Text: "hidden"}},
		}},
	}
	if got := inlineChanges(binaryWithHunks); got != nil {
		t.Fatalf("binary with hunks = %+v, want nil", got)
	}

	if got := inlineChanges(changedFile{}); got != nil {
		t.Fatalf("zero file = %+v, want nil", got)
	}

	if _, ok := inlineChangeAt(nil, 1); ok {
		t.Fatal("nil map returned a change")
	}
}

// checkInlineLine asserts one stored line's kind and replaced text.
func checkInlineLine(t *testing.T, changes map[int]inlineLine, line int, kind inlineKind, oldText string) {
	t.Helper()

	change, ok := inlineChangeAt(changes, line)
	if !ok {
		t.Fatalf("line %d: no change", line)
	}

	if change.Kind != kind || change.OldText != oldText {
		t.Fatalf("line %d = %+v, want kind %q old %q", line, change, kind, oldText)
	}
}
