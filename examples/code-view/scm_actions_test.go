package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestSCMContextMenuFileRows(t *testing.T) {
	app := newTestApp(t)
	app.mode = "changes"
	app.refresh()

	if len(app.changes.Files) == 0 {
		t.Fatal("sample has no changes")
	}

	if len(app.changes.Staged) == 0 || len(app.changes.Unstaged) == 0 {
		t.Fatalf("sample split = staged %d unstaged %d, want both", len(app.changes.Staged), len(app.changes.Unstaged))
	}

	unstagedPath := app.changes.Unstaged[0].Path
	for _, id := range []string{"change-" + unstagedPath, "stage-" + unstagedPath, "discard-" + unstagedPath} {
		rows := app.ContextMenuFor(id)
		if len(rows) != 5 {
			t.Fatalf("%s rows = %d, want 5", id, len(rows))
		}

		if rows[0].Label != "Stage Changes" || rows[1].Label != "Discard Changes" {
			t.Fatalf("%s labels = %q %q", id, rows[0].Label, rows[1].Label)
		}

		if rows[2].Label != "Stage All" || rows[4].Label != "Discard All" {
			t.Fatalf("%s bulk labels = %q %q", id, rows[2].Label, rows[4].Label)
		}
	}

	stagedPath := app.changes.Staged[0].Path
	for _, id := range []string{"staged-" + stagedPath, "unstage-" + stagedPath, "discard-staged-" + stagedPath} {
		rows := app.ContextMenuFor(id)
		if len(rows) == 0 || rows[0].Label != "Unstage Changes" {
			t.Fatalf("%s first row = %+v, want Unstage Changes", id, rows)
		}
	}
}

func TestSCMContextMenuBulk(t *testing.T) {
	app := newTestApp(t)
	app.mode = "changes"
	app.refresh()

	for _, id := range []string{"changes-list", "changes-head", "history-working", "changes-staged-head", "changes-unstaged-head"} {
		rows := app.ContextMenuFor(id)
		if len(rows) != 3 {
			t.Fatalf("%s rows = %d, want 3", id, len(rows))
		}

		if rows[0].Label != "Stage All" || rows[1].Label != "Unstage All" || rows[2].Label != "Discard All" {
			t.Fatalf("%s labels = %q %q %q", id, rows[0].Label, rows[1].Label, rows[2].Label)
		}
	}
}

func TestStagedSplitRenders(t *testing.T) {
	app := newTestApp(t)
	app.mode = "changes"
	app.refresh()

	view := app.View()
	if !view.Diff.HasStaged || !view.Diff.HasUnstaged {
		t.Fatalf("sections = staged %v unstaged %v", view.Diff.HasStaged, view.Diff.HasUnstaged)
	}

	if view.Diff.StagedCount != len(app.changes.Staged) || view.Diff.UnstagedCount != len(app.changes.Unstaged) {
		t.Fatalf("counts = %d/%d, want %d/%d", view.Diff.StagedCount, view.Diff.UnstagedCount, len(app.changes.Staged), len(app.changes.Unstaged))
	}

	if len(view.Diff.Staged) == 0 || len(view.Diff.Unstaged) == 0 {
		t.Fatalf("visible rows = staged %d unstaged %d", len(view.Diff.Staged), len(view.Diff.Unstaged))
	}

	if !view.Diff.Staged[0].Staged || view.Diff.Unstaged[0].Staged {
		t.Fatal("section flags are swapped")
	}
}

func TestMultiSelectCtrlAAndClick(t *testing.T) {
	app := newTestApp(t)
	app.mode = "changes"
	app.refresh()

	ctx := context.Background()
	if err := app.onKeyDown(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	// Ctrl+A selects every working-tree path.
	if err := app.onKeyDown(ctx, "a"); err != nil {
		t.Fatal(err)
	}

	if err := app.onKeyUp(ctx, "control"); err != nil {
		t.Fatal(err)
	}

	if got, want := app.selectedCount(), len(app.workingAllPaths()); got != want {
		t.Fatalf("selected = %d, want %d", got, want)
	}

	// Plain click collapses back to one file.
	app.ctrl = false
	second := app.workingAllPaths()[1]
	if err := app.onClick(ctx, ownframe.Box{ID: "change-" + second, Action: "diff-file:" + second}); err != nil {
		t.Fatal(err)
	}

	if got := app.selectedCount(); got != 1 {
		t.Fatalf("after click selected = %d, want 1", got)
	}

	// Ctrl-click adds a second file without clearing the first.
	app.ctrl = true
	third := app.workingAllPaths()[2]
	if err := app.onClick(ctx, ownframe.Box{ID: "change-" + third, Action: "diff-file:" + third}); err != nil {
		t.Fatal(err)
	}
	app.ctrl = false

	if got := app.selectedCount(); got != 2 {
		t.Fatalf("after ctrl-click selected = %d, want 2", got)
	}

	if !app.View().Diff.Unstaged[0].Multi && !app.View().Diff.Staged[0].Multi {
		// At least one visible row must carry the multi flag.
		found := false
		for _, row := range append(app.View().Diff.Staged, app.View().Diff.Unstaged...) {
			if row.Multi {
				found = true
				break
			}
		}
		if !found {
			t.Fatal("no visible row shows the multi-select state")
		}
	}
}

func TestDiscardConfirmFlow(t *testing.T) {
	dir, _ := historyRepo(t)
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.mode = "changes"
	app.refresh()

	if len(app.workingAllPaths()) == 0 {
		t.Fatal("live repo has no changes")
	}

	path := app.workingAllPaths()[0]
	app.openDiscardConfirm("discard-file", []string{path})

	if !app.View().Confirm.Show {
		t.Fatal("confirm did not open")
	}

	if !strings.Contains(app.View().Confirm.Title, "Discard") || !strings.Contains(app.View().Confirm.Message, "Do you really want") {
		t.Fatalf("confirm copy = %+v", app.View().Confirm)
	}

	if err := app.onKeyDown(ctx, "escape"); err != nil {
		t.Fatal(err)
	}

	if app.View().Confirm.Show {
		t.Fatal("escape did not close confirm")
	}

	// Menu discard opens confirm instead of deleting.
	app.openDiscardConfirm("discard-file", []string{path})
	if err := app.MenuAction(ctx, "code-view:discard-file:"+path); err != nil {
		t.Fatal(err)
	}

	if !app.View().Confirm.Show {
		t.Fatal("menu discard did not open confirm")
	}

	// A real press then click on Cancel must close the dialog. A consumed
	// press on the same id makes Click redraw only, so the dialog buttons
	// must not consume the press.
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	cancel, ok := boxByID(app.Page().Boxes(), "confirm-cancel")
	if !ok {
		t.Fatal("missing confirm-cancel")
	}

	x, y := cancel.X+cancel.W/2, cancel.Y+cancel.H/2
	if err := app.Page().Edit().Press(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	if err := app.Page().Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	if app.View().Confirm.Show {
		t.Fatal("press and click on Cancel did not close confirm")
	}
}

func TestHistoryContextMenuCopy(t *testing.T) {
	ownframe.UseMemory(true)
	defer ownframe.UseMemory(false)

	app, hashes := newHistoryApp(t)
	if len(hashes) == 0 {
		t.Fatal("no history")
	}

	app.mode = "changes"
	app.refresh()

	var short string
	for i := range app.commits.Commits {
		if app.commits.Commits[i].Hash == hashes[0] {
			short = app.commits.Commits[i].Short
		}
	}

	if short == "" {
		t.Fatal("commit short not found")
	}

	rows := app.ContextMenuFor("commit-" + short)
	if len(rows) != 2 {
		t.Fatalf("history rows = %d, want 2", len(rows))
	}

	if rows[0].Label != "Copy Message" || rows[1].Label != "Copy Commit ID" {
		t.Fatalf("history labels = %q %q", rows[0].Label, rows[1].Label)
	}

	if strings.HasPrefix(rows[0].ID, "code-view:discard") || strings.HasPrefix(rows[0].ID, "code-view:stage") {
		t.Fatalf("history menu shows working-tree rows: %q", rows[0].ID)
	}

	ctx := context.Background()
	if err := app.MenuAction(ctx, rows[1].ID); err != nil {
		t.Fatal(err)
	}

	if got, _ := ownframe.Read(); got != hashes[0] {
		t.Fatalf("clipboard ID = %q, want %q", got, hashes[0])
	}

	if err := app.MenuAction(ctx, rows[0].ID); err != nil {
		t.Fatal(err)
	}

	if got, _ := ownframe.Read(); got == "" || !strings.Contains(got, "print hello") {
		t.Fatalf("clipboard message = %q", got)
	}
}

func TestDiscardFileLive(t *testing.T) {
	dir, _ := historyRepo(t)
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)

	if len(app.changes.Files) == 0 {
		t.Fatal("live repo has no working-tree change")
	}

	path := app.changes.Files[0].Path
	full := filepath.Join(dir, filepath.FromSlash(path))
	before, err := os.ReadFile(full)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(before), "working change") {
		t.Fatalf("unexpected bytes before discard: %q", before)
	}

	if err := app.discardFile(ctx, path); err != nil {
		t.Fatal(err)
	}

	after, err := os.ReadFile(full)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(after), "working change") {
		t.Fatalf("discard kept the edit: %q", after)
	}

	if findFile(app.changes.Files, path) != nil {
		t.Fatalf("discarded file still listed: %v", app.changes.Files)
	}
}

func TestDiscardUntrackedLive(t *testing.T) {
	dir, _ := historyRepo(t)
	writeFile(t, dir, "scratch.txt", "scratch\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)

	if findFile(app.changes.Files, "scratch.txt") == nil {
		t.Fatal("untracked file not listed")
	}

	if err := app.discardFile(ctx, "scratch.txt"); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(dir, "scratch.txt")); !os.IsNotExist(err) {
		t.Fatalf("untracked file still on disk: %v", err)
	}
}

func TestSampleGuardBlocksWrites(t *testing.T) {
	app := newTestApp(t)
	if !app.ws.Sample {
		t.Fatal("test app is not the sample")
	}

	ctx := context.Background()
	if err := app.discardFile(ctx, app.changes.Files[0].Path); err != nil {
		t.Fatal(err)
	}

	if !app.View().Toast.Show {
		t.Fatal("sample discard shows no toast")
	}
}
