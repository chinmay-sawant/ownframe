package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestSliceSegments(t *testing.T) {
	t.Parallel()

	segs := []segment{
		{Text: "hello ", Class: classPlain},
		{Text: "`code`", Class: classCode},
		{Text: " world", Class: classPlain},
	}
	join := func(ss []segment) string {
		var b strings.Builder
		for _, s := range ss {
			b.WriteString(s.Text)
		}
		return b.String()
	}

	if got := join(sliceSegments(segs, 0, 100)); got != "hello `code` world" {
		t.Fatalf("full window = %q", got)
	}

	if got := join(sliceSegments(segs, 0, 5)); got != "hello" {
		t.Fatalf("head = %q", got)
	}

	if got := join(sliceSegments(segs, 6, 6)); got != "`code`" {
		t.Fatalf("middle = %q", got)
	}

	if got := join(sliceSegments(segs, 12, 100)); got != " world" {
		t.Fatalf("tail = %q", got)
	}

	if got := sliceSegments(segs, 100, 10); len(got) != 0 {
		t.Fatalf("past-end slice has %d runs", len(got))
	}

	if got := sliceSegments(segs, 0, 0); len(got) != 0 {
		t.Fatalf("empty slice has %d runs", len(got))
	}

	// The source slice keeps its text: slicing must not rewrite runs.
	if got := join(segs); got != "hello `code` world" {
		t.Fatalf("source rewritten to %q", got)
	}
}

// longMarkdownLine builds a single markdown source line with many inline
// spans, the shape that used to overflow its line box and overlap.
func longMarkdownLine() string {
	var b strings.Builder
	b.WriteString("A press held within 8 px for 450 ms calls the inner screen's `LongPress`.")
	for i := 0; i < 12; i++ {
		b.WriteString(" Any other box reaches `Handlers.LongPress`, which receives `page.Click` finds it.")
	}
	return b.String()
}

func TestUnwrappedCodeRowsFitViewport(t *testing.T) {
	dir := t.TempDir()
	line := longMarkdownLine()
	if err := os.WriteFile(filepath.Join(dir, "doc.md"), []byte("# T\n\n"+line+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("x\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	app.openFile("doc.md")
	if app.wrap {
		app.toggleWrap()
	}
	if app.wrap {
		t.Fatal("wrap still on")
	}
	cols := app.codeColsForGroup(0)
	cellW := app.zf(codeCellW)

	for _, scrollX := range []int{0, 50} {
		app.SetScrollX(scrollX)
		if err := app.Page().Redraw(ctx); err != nil {
			t.Fatal(err)
		}
		wantCol := int(float64(scrollX) / cellW)
		wantRem := scrollX - int(float64(wantCol)*cellW)
		for _, row := range app.View().Rows {
			if got := segmentRuneLen(row.Segments); got > cols {
				t.Fatalf("scrollX=%d: row %d lays %d runes, budget %d", scrollX, row.Line, got, cols)
			}
			if row.ScrollX != wantRem {
				t.Fatalf("scrollX=%d: row offset = %d, want %d", scrollX, row.ScrollX, wantRem)
			}
		}
		// The painted text is the source window the scroll position asks for.
		found := false
		for _, row := range app.View().Rows {
			if row.Line != 2 || len(row.Segments) == 0 {
				continue
			}
			found = true
			var b strings.Builder
			for _, s := range row.Segments {
				b.WriteString(s.Text)
			}
			want := string([]rune(line)[wantCol:min(wantCol+cols, len([]rune(line)))])
			if b.String() != want {
				t.Fatalf("scrollX=%d: painted %q, want %q", scrollX, b.String(), want)
			}
		}
		if !found {
			t.Fatalf("scrollX=%d: long line missing from view", scrollX)
		}
	}
}

func TestUnwrappedDiffRowsFitViewport(t *testing.T) {
	dir := t.TempDir()
	oldLine := longMarkdownLine()
	newLine := strings.Replace(oldLine, "450 ms", "about 450 ms", 1)
	if err := os.WriteFile(filepath.Join(dir, "doc.md"), []byte("# T\n\n"+oldLine+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	runGit := func(args ...string) {
		t.Helper()
		c := exec.CommandContext(context.Background(), "git", args...) //nolint:gosec // reason: test helper runs the fixed git binary with a fixed subcommand list on a temp repo
		c.Dir = dir
		c.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "HOME="+dir)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	runGit("init", "-q")
	runGit("config", "user.email", "t@t.t")
	runGit("config", "user.name", "t")
	runGit("add", ".")
	runGit("commit", "-qm", "init")
	if err := os.WriteFile(filepath.Join(dir, "doc.md"), []byte("# T\n\n"+newLine+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	app.reloadChanges(ctx, false)
	if len(app.changes.Files) == 0 {
		t.Fatal("no changes loaded")
	}
	app.SetDevTools(false)
	if err := app.onClick(ctx, ownframe.Box{Action: "mode-changes"}); err != nil {
		t.Fatal(err)
	}
	if app.wrap {
		app.toggleWrap()
	}
	unified, _ := app.diffColumns()
	cellW := app.zf(diffCellW)
	app.SetScrollX(50)
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	wantCol := int(float64(50) / cellW)
	wantRem := 50 - int(float64(wantCol)*cellW)
	seen := 0
	for _, row := range app.View().Diff.Rows {
		if row.Kind == rowExpand || row.Kind == rowHunk || row.Kind == "meta" {
			continue
		}
		if len(row.Left) == 0 {
			continue
		}
		seen++
		if got := segmentRuneLen(row.Left); got > unified {
			t.Fatalf("row kind %s lays %d runes, budget %d", row.Kind, got, unified)
		}
		if row.ScrollX != wantRem {
			t.Fatalf("row offset = %d, want %d", row.ScrollX, wantRem)
		}
		_ = wantCol
	}
	if seen == 0 {
		t.Fatal("no diff text rows in view")
	}
}
