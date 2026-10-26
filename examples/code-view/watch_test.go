package main

import (
	"context"
	"testing"
	"time"
)

func TestWatcherUpdatesWorkingTree(t *testing.T) {
	dir, _ := historyRepo(t)

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.reloadChanges(context.Background(), false)
	app.mode = "changes"

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.StartWatcher(ctx); err != nil {
		t.Fatal(err)
	}

	defer app.StopWatcher()

	writeFile(t, dir, "app.go", "package main\n\nfunc main() {\n\tprintln(2)\n}\n")

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		if err := app.tickWatch(context.Background()); err != nil {
			t.Fatal(err)
		}

		if app.View().Changed == 2 && app.View().Diff.Path == "README.md" {
			return
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("watcher did not update: changed=%d diff=%q", app.View().Changed, app.View().Diff.Path)
}

func TestWatcherSkipsSample(t *testing.T) {
	app := newTestApp(t)

	if err := app.StartWatcher(context.Background()); err != nil {
		t.Fatal(err)
	}

	if app.watch != nil {
		t.Fatal("sample workspace started a watcher")
	}
}

func TestApplyWatchChangesPreservesSelection(t *testing.T) {
	app, _ := newHistoryApp(t)
	app.diffPath = "README.md"
	app.diffSource = "working"
	app.changePage = 0

	changes := app.changes
	changes.Files = append([]changedFile(nil), changes.Files...)

	app.applyWatchChanges(context.Background(), changes)

	if app.View().Diff.Path != "README.md" {
		t.Fatalf("selected diff moved to %q", app.View().Diff.Path)
	}

	if app.View().Diff.Source != "working" {
		t.Fatalf("diff source = %q", app.View().Diff.Source)
	}
}
