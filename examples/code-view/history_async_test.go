package main

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"
)

const (
	historyAsyncCommits     = 26
	historyAsyncScrollRows  = 100
	historyAsyncScrollLimit = 100 * time.Millisecond
	historyAsyncWait        = 3 * time.Second
)

// newHistoryAsyncRepo builds a temp repository with exactly 26 commits: 24 on
// main, one on a side branch, and a merge, so the history graph has more than
// one lane to draw.
func newHistoryAsyncRepo(t *testing.T) string {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	runGit(t, dir, "config", "user.name", "Code View Test")
	runGit(t, dir, "config", "user.email", "code-view@example.com")

	writeFile(t, dir, "README.md", "# History async\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "commit 01")

	// 23 more on main brings the branch to 24 commits.
	for i := 2; i <= 24; i++ {
		writeFile(t, dir, "app.go", fmt.Sprintf("package main\n\nfunc main() { println(%d) }\n", i))
		runGit(t, dir, "add", "-A")
		runGit(t, dir, "commit", "-q", "-m", fmt.Sprintf("commit %02d", i))
	}

	runGit(t, dir, "checkout", "-q", "-b", "side")
	writeFile(t, dir, "side.go", "package main\n")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "commit", "-q", "-m", "side commit")

	runGit(t, dir, "checkout", "-q", "main")
	runGit(t, dir, "merge", "-q", "--no-ff", "side", "-m", "merge side")

	if count := runGit(t, dir, "rev-list", "--count", "HEAD"); count != fmt.Sprint(historyAsyncCommits) {
		t.Fatalf("repo has %s commits, want %d", count, historyAsyncCommits)
	}

	return dir
}

// TestHistoryAsync proves the history rail never runs git on the scroll call:
// the page marks a load in flight, the scroll returns at once, and a later tick
// installs the larger log with graph cells.
func TestHistoryAsync(t *testing.T) {
	dir := newHistoryAsyncRepo(t)

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	app.reloadChanges(ctx, false)
	app.setMode(ctx, "changes")

	if got := len(app.commits.Commits); got != maxCommits {
		t.Fatalf("initial commits = %d, want %d", got, maxCommits)
	}

	visible := app.historyVisibleRows()
	if visible >= len(app.commits.Commits) {
		t.Fatalf("history rail shows every commit: visible=%d total=%d", visible, len(app.commits.Commits))
	}

	before := len(app.commits.Commits)
	started := time.Now()
	app.scrollHistory(ctx, historyAsyncScrollRows)
	elapsed := time.Since(started)
	loading := app.historyScroll.Loading

	t.Logf("scrollHistory(%d rows) returned in %v, loading=%v", historyAsyncScrollRows, elapsed, loading)

	if elapsed >= historyAsyncScrollLimit {
		t.Fatalf("scrollHistory blocked on the history load for %v", elapsed)
	}

	deadline := time.Now().Add(historyAsyncWait)
	for time.Now().Before(deadline) && len(app.commits.Commits) <= before {
		if err := app.tickWatch(ctx); err != nil {
			t.Fatal(err)
		}

		time.Sleep(10 * time.Millisecond)
	}

	loaded := len(app.commits.Commits)
	grown := loaded > before
	if !loading && !grown {
		t.Fatalf("history load never started: loading=%v commits=%d", loading, loaded)
	}

	if !grown {
		t.Fatalf("history did not load more commits within %v: commits=%d", historyAsyncWait, loaded)
	}

	if loading {
		t.Logf("historyScroll.Loading was true right after the scroll; it cleared when the result arrived")
	}

	if app.historyScroll.Loading {
		t.Fatalf("history load still in flight after the commit count grew to %d", loaded)
	}

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	rows := app.View().History.Commits
	if len(rows) == 0 {
		t.Fatal("history rail has no visible commits after the load")
	}

	lanes := 0

	for _, row := range rows {
		if len(row.Graph) == 0 {
			t.Fatalf("commit %s has no graph cells", row.Short)
		}

		if row.Lanes < 1 || row.Lanes > maxCommitLanes {
			t.Fatalf("commit %s lanes = %d, want 1..%d", row.Short, row.Lanes, maxCommitLanes)
		}

		for _, cell := range row.Graph {
			if cell.Lane < 0 || cell.Lane >= row.Lanes {
				t.Fatalf("commit %s cell lane = %d, row lanes = %d", row.Short, cell.Lane, row.Lanes)
			}
		}

		lanes = max(lanes, row.Lanes)
	}

	t.Logf("commits grew from %d to %d; visible rows=%d; widest lanes=%d", before, loaded, len(rows), lanes)
}
