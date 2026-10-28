package main

import (
	"context"
	"os"
	"path/filepath"
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

func TestWatcherRescanAddsAndRemovesFiles(t *testing.T) {
	dir, _ := historyRepo(t)

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.folders["pkg"] = true

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.StartWatcher(ctx); err != nil {
		t.Fatal(err)
	}

	defer app.StopWatcher()

	writeFile(t, dir, "pkg/new.go", "package pkg\n")

	waitWatch(t, app, func() bool {
		return app.ws.ByPath["pkg/new.go"] != nil && app.View().WatchNote != ""
	})

	if !app.folders["pkg"] {
		t.Fatal("rescan dropped an expanded folder")
	}

	if err := os.Remove(filepath.Join(dir, "pkg", "new.go")); err != nil {
		t.Fatal(err)
	}

	waitWatch(t, app, func() bool {
		return app.ws.ByPath["pkg/new.go"] == nil
	})
}

func TestWatcherSeesExternalBranchChange(t *testing.T) {
	dir, _ := historyRepo(t)

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.StartWatcher(ctx); err != nil {
		t.Fatal(err)
	}

	defer app.StopWatcher()

	runGit(t, dir, "checkout", "-q", "-b", "feature")

	waitWatch(t, app, func() bool {
		return app.View().Branch == "feature" || app.headBranch == "feature"
	})
}

func waitWatch(t *testing.T, app *App, ready func() bool) {
	t.Helper()

	deadline := time.Now().Add(6 * time.Second)

	for time.Now().Before(deadline) {
		if err := app.tickWatch(context.Background()); err != nil {
			t.Fatal(err)
		}

		if ready() {
			return
		}

		time.Sleep(30 * time.Millisecond)
	}

	t.Fatal("watcher did not deliver the expected state")
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

func TestWatcherPeriodicPollingDetectsChangesWithoutFSNotify(t *testing.T) {
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

	// Disarm fsnotify by removing all watched paths from the watcher.
	// This ensures no fsnotify events can fire, so any detected update
	// must originate from background periodic git diff polling.
	for _, path := range app.watch.WatchList() {
		_ = app.watch.Remove(path)
	}

	// Verify optimization: clean repository does not push spurious updates across ticks.
	time.Sleep(2100 * time.Millisecond)
	select {
	case u := <-app.watchUpdates:
		t.Fatalf("unexpected spurious watch update during clean state: %+v", u)
	default:
	}

	// Modify a file in the working tree without any fsnotify event.
	writeFile(t, dir, "app.go", "package main\n\nfunc main() {\n\tprintln(\"polled change\")\n}\n")

	waitWatch(t, app, func() bool {
		return app.View().Changed == 2 && findFile(app.changes.Files, "app.go") != nil
	})
}

func TestWatcherPeriodicPollingDetectsBranchChangeWithoutFSNotify(t *testing.T) {
	dir, _ := historyRepo(t)

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.reloadChanges(context.Background(), false)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := app.StartWatcher(ctx); err != nil {
		t.Fatal(err)
	}
	defer app.StopWatcher()

	// Disarm fsnotify completely.
	for _, path := range app.watch.WatchList() {
		_ = app.watch.Remove(path)
	}

	runGit(t, dir, "checkout", "-q", "-b", "polled-branch")

	waitWatch(t, app, func() bool {
		return app.View().Branch == "polled-branch" || app.headBranch == "polled-branch"
	})
}

func TestChangeSetDelta(t *testing.T) {
	base := changeSet{
		Branch: "main",
		Note:   "",
		Files: []changedFile{
			{
				Path:   "main.go",
				Status: "M",
				Add:    1,
				Del:    0,
				Hunks: []hunk{
					{
						OldStart: 1, OldCount: 1, NewStart: 1, NewCount: 2,
						Header: "@@ -1,1 +1,2 @@",
						Lines: []diffLine{
							{Kind: ' ', Text: "package main"},
							{Kind: '+', Text: "func init() {}"},
						},
					},
				},
			},
		},
	}

	// Identical change set has no delta.
	if changeSetDelta(base, base) {
		t.Fatal("expected no delta for identical changeSets")
	}

	// Different branch.
	otherBranch := base
	otherBranch.Branch = "dev"
	if !changeSetDelta(base, otherBranch) {
		t.Fatal("expected delta for different branch")
	}

	// Different note.
	otherNote := base
	otherNote.Note = "Working tree clean"
	if !changeSetDelta(base, otherNote) {
		t.Fatal("expected delta for different note")
	}

	// Different file count.
	fewerFiles := base
	fewerFiles.Files = nil
	if !changeSetDelta(base, fewerFiles) {
		t.Fatal("expected delta for fewer files")
	}

	// Different file status.
	otherStatus := base
	otherStatus.Files = []changedFile{{Path: "main.go", Status: "A", Add: 1, Del: 0}}
	if !changeSetDelta(base, otherStatus) {
		t.Fatal("expected delta for different status")
	}

	// Different hunk line content.
	otherHunk := base
	otherHunk.Files = []changedFile{
		{
			Path:   "main.go",
			Status: "M",
			Add:    1,
			Del:    0,
			Hunks: []hunk{
				{
					OldStart: 1, OldCount: 1, NewStart: 1, NewCount: 2,
					Header: "@@ -1,1 +1,2 @@",
					Lines: []diffLine{
						{Kind: ' ', Text: "package main"},
						{Kind: '+', Text: "func init() { panic(1) }"},
					},
				},
			},
		},
	}
	if !changeSetDelta(base, otherHunk) {
		t.Fatal("expected delta for different hunk lines")
	}
}
