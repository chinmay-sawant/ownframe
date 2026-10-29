package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDiffMarksOversizedBytes(t *testing.T) {
	t.Parallel()

	line := strings.Repeat("x", maxDiffBytes+1)
	raw := "diff --git a/big.js b/big.js\n" +
		"--- a/big.js\n" +
		"+++ b/big.js\n" +
		"@@ -1,1 +1,1 @@\n" +
		"-" + line + "\n" +
		"+small\n"

	files := parseDiff(raw)
	if len(files) != 1 {
		t.Fatalf("files = %d, want 1", len(files))
	}

	file := files[0]
	if !file.Oversized {
		t.Fatalf("file not oversized: bytes=%d lines=%d", file.DiffBytes, file.DiffLines)
	}

	if file.DiffBytes <= maxDiffBytes {
		t.Fatalf("DiffBytes = %d, want over %d", file.DiffBytes, maxDiffBytes)
	}

	rows := diffRowsUnified(file, langJS)
	if len(rows) != 1 || rows[0].Kind != rowMeta {
		t.Fatalf("unified rows = %+v, want one meta row", rows)
	}

	if !strings.Contains(rows[0].Hunk, "Diff skipped") || !strings.Contains(rows[0].Hunk, "512 KB") {
		t.Fatalf("limit note = %q", rows[0].Hunk)
	}

	if split := diffRowsSplit(file, langJS); len(split) != 1 || split[0].Kind != rowMeta {
		t.Fatalf("split rows = %+v, want one meta row", split)
	}
}

func TestParseDiffMarksOversizedLines(t *testing.T) {
	t.Parallel()

	var raw strings.Builder
	raw.WriteString("diff --git a/many.go b/many.go\n")
	raw.WriteString("--- a/many.go\n")
	raw.WriteString("+++ b/many.go\n")
	raw.WriteString("@@ -0,0 +1,")
	raw.WriteString("20001")
	raw.WriteString(" @@\n")

	for range maxDiffLines + 1 {
		raw.WriteString("+x\n")
	}

	files := parseDiff(raw.String())
	if len(files) != 1 {
		t.Fatalf("files = %d, want 1", len(files))
	}

	file := files[0]
	if !file.Oversized || file.DiffLines != maxDiffLines+1 {
		t.Fatalf("file = bytes %d lines %d oversized %v", file.DiffBytes, file.DiffLines, file.Oversized)
	}

	if rows := diffRowsUnified(file, langGo); len(rows) != 1 || rows[0].Kind != rowMeta {
		t.Fatalf("rows = %+v, want one meta row", rows)
	}
}

func TestUntrackedChangeIsMeasured(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("first\nsecond\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	file, ok := untrackedChange(dir, "notes.txt")
	if !ok {
		t.Fatal("untracked file skipped")
	}

	if file.DiffLines != 2 || file.DiffBytes == 0 || file.Oversized {
		t.Fatalf("measure = bytes %d lines %d oversized %v", file.DiffBytes, file.DiffLines, file.Oversized)
	}
}

func TestLiveChangesMarksOversizedDeletion(t *testing.T) {
	dir := t.TempDir()

	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.name", "Code View Test")
	runGit(t, dir, "config", "user.email", "code-view@example.com")

	writeFile(t, dir, "bundle.js", strings.Repeat("x", maxDiffBytes+1)+"\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "add bundle")

	if err := os.Remove(filepath.Join(dir, "bundle.js")); err != nil {
		t.Fatal(err)
	}

	changes, err := loadLiveChanges(t.Context(), dir)
	if err != nil {
		t.Fatal(err)
	}

	file := findFile(changes.Files, "bundle.js")
	if file == nil {
		t.Fatal("deleted bundle missing from the working tree diff")
	}

	if !file.Oversized || file.DiffBytes <= maxDiffBytes {
		t.Fatalf("deletion measure = bytes %d lines %d oversized %v", file.DiffBytes, file.DiffLines, file.Oversized)
	}
}

func TestCommitPatchMarksTruncatedRead(t *testing.T) {
	dir := t.TempDir()

	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.name", "Code View Test")
	runGit(t, dir, "config", "user.email", "code-view@example.com")

	writeFile(t, dir, "bundle.js", strings.Repeat("x", maxCommitPatchBytes+1)+"\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "add bundle")
	hash := runGit(t, dir, "rev-parse", "HEAD")

	file, err := loadCommitFile(t.Context(), dir, hash, "bundle.js")
	if err != nil {
		t.Fatal(err)
	}

	if !file.DiffTruncated || !file.Oversized {
		t.Fatalf("commit patch = bytes %d lines %d truncated %v oversized %v", file.DiffBytes, file.DiffLines, file.DiffTruncated, file.Oversized)
	}

	rows := diffRowsUnified(*file, langJS)
	if len(rows) != 1 || rows[0].Kind != rowMeta {
		t.Fatalf("rows = %+v, want one meta row", rows)
	}

	if !strings.Contains(rows[0].Hunk, "read limit") {
		t.Fatalf("limit row = %q", rows[0].Hunk)
	}
}
