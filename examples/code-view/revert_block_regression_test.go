package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRevertBlockPointer(t *testing.T) {
	for _, split := range []bool{false, true} {
		t.Run(strconv.FormatBool(split), func(t *testing.T) {
			dir, _ := historyRepo(t)
			writeFile(t, dir, "README.md", "one\ntwo\nthree\nfour\nfive\n")
			runGit(t, dir, "add", "README.md")
			runGit(t, dir, "commit", "-q", "-m", "base")
			writeFile(t, dir, "README.md", "ONE\ntwo\nthree\nFOUR\nfive\n")
			app, err := New(dir)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			app.reloadChanges(ctx, false)
			app.mode = "changes"
			app.selectDiff("README.md")
			app.setSplit(split)
			app.wrap = false
			app.refresh()
			if err := app.Page().Redraw(ctx); err != nil {
				t.Fatal(err)
			}
			box, ok := firstRevertLineBox(app)
			if !ok {
				t.Fatal("missing revert button")
			}
			if err := app.Page().Edit().Press(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
				t.Fatal(err)
			}
			if err := app.Page().Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(filepath.Join(dir, "README.md")) //nolint:gosec // path is inside a t.TempDir fixture
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != "one\ntwo\nthree\nFOUR\nfive\n" {
				t.Fatalf("file after click: %q", got)
			}
			if app.mode != "changes" {
				t.Fatalf("mode after click: %s", app.mode)
			}
			if strings.Contains(app.View().Diff.EmptyTitle, "unavailable") {
				t.Fatal("diff unavailable")
			}
		})
	}
}

func TestSplitRevertPointer(t *testing.T) {
	dir, _ := historyRepo(t)
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.mode = "changes"
	app.selectDiff("README.md")
	app.setSplit(true)
	app.wrap = false
	app.refresh()
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	box, ok := firstRevertLineBox(app)
	if !ok {
		t.Fatal("missing revert button")
	}
	if err := app.Page().Edit().Press(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}
	if err := app.Page().Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "README.md")) //nolint:gosec // path is inside a t.TempDir fixture
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "working change") {
		t.Fatalf("split pointer did not revert: %q", got)
	}
}

func TestRevertBlockStagedUndoAndRefresh(t *testing.T) {
	for _, staged := range []bool{false, true} {
		t.Run(strconv.FormatBool(staged), func(t *testing.T) {
			dir, _ := historyRepo(t)
			base := "one\ntwo\nthree\nfour\nfive\n"
			edited := "ONE\ntwo\nthree\nFOUR\nfive\n"
			writeFile(t, dir, "README.md", base)
			runGit(t, dir, "add", "README.md")
			runGit(t, dir, "commit", "-qm", "base")
			writeFile(t, dir, "README.md", edited)
			if staged {
				runGit(t, dir, "add", "README.md")
			}
			app, err := New(dir)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			app.openFile("README.md")
			app.reloadChanges(ctx, false)
			app.selectDiff("README.md")
			if err := app.Page().Redraw(ctx); err != nil {
				t.Fatal(err)
			}
			box, ok := firstRevertLineBox(app)
			if !ok {
				t.Fatal("no revert")
			}
			if err := app.Page().Edit().Press(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
				t.Fatal(err)
			}
			if err := app.Page().Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
				t.Fatal(err)
			}
			want := "one\ntwo\nthree\nFOUR\nfive\n"
			check := func(want string) {
				t.Helper()
				got, err := os.ReadFile(filepath.Join(dir, "README.md")) //nolint:gosec // path is inside a t.TempDir fixture
				if err != nil {
					t.Fatal(err)
				}
				if string(got) != want {
					t.Fatalf("disk=%q want %q", got, want)
				}
				if app.mode != "changes" {
					t.Fatalf("redirected to %s", app.mode)
				}
				if got := strings.Join(app.groupDoc(0).Lines, "\n"); strings.TrimSuffix(got, "\n") != strings.TrimSuffix(want, "\n") {
					t.Fatalf("stale document=%q", got)
				}
			}
			check(want)
			if staged {
				if got := runGit(t, dir, "show", ":README.md"); strings.TrimSpace(got) != strings.TrimSpace(want) {
					t.Fatalf("index=%q", got)
				}
			}
			if err := app.undoRevertHunk(ctx); err != nil {
				t.Fatal(err)
			}
			check(edited)
			if err := app.redoRevertHunk(ctx); err != nil {
				t.Fatal(err)
			}
			check(want)
			if err := app.Page().Redraw(ctx); err != nil {
				t.Fatal(err)
			}
			box, ok = firstRevertLineBox(app)
			if !ok {
				t.Fatal("remaining block missing")
			}
			if err := app.Page().Edit().Press(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
				t.Fatal(err)
			}
			if err := app.Page().Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
				t.Fatal(err)
			}
			check(base)
			if app.View().Diff.HasFiles {
				t.Fatal("clean diff still shows file")
			}
		})
	}
}

func TestRevertBlockInsertDelete(t *testing.T) {
	for _, tc := range []struct{ name, base, edited, want string }{
		{"insert", "a\nb\nc\nd\ne\n", "a\ninsert\nb\nc\nD\ne\n", "a\nb\nc\nD\ne\n"},
		{"delete", "a\nb\nc\nd\ne\n", "b\nc\nD\ne\n", "a\nb\nc\nD\ne\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, _ := historyRepo(t)
			writeFile(t, dir, "README.md", tc.base)
			runGit(t, dir, "add", "README.md")
			runGit(t, dir, "commit", "-qm", "base")
			writeFile(t, dir, "README.md", tc.edited)
			app, err := New(dir)
			if err != nil {
				t.Fatal(err)
			}
			ctx := context.Background()
			app.reloadChanges(ctx, false)
			app.selectDiff("README.md")
			if err := app.Page().Redraw(ctx); err != nil {
				t.Fatal(err)
			}
			box, ok := firstRevertLineBox(app)
			if !ok {
				t.Fatal("no revert")
			}
			if err := app.Page().Edit().Press(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
				t.Fatal(err)
			}
			if err := app.Page().Click(ctx, box.X+box.W/2, box.Y+box.H/2); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(filepath.Join(dir, "README.md")) //nolint:gosec // path is inside a t.TempDir fixture
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
			if err := app.undoRevertHunk(ctx); err != nil {
				t.Fatal(err)
			}
			got, err = os.ReadFile(filepath.Join(dir, "README.md")) //nolint:gosec // path is inside a t.TempDir fixture
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.edited {
				t.Fatalf("undo=%q want %q", got, tc.edited)
			}
		})
	}
}
