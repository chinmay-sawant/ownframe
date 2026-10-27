package main

import (
	"context"
	"fmt"
	"testing"
)

// TestHistoryScrollE2E drives the History rail through the Page: wheel over the
// pane, drag the thumb, and reach the last commit. The sample workspace keeps
// the history fetch on its bundled fixture instead of running git.
func TestHistoryScrollE2E(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	app.commits = historyScrollCommits(30)
	app.historyScroll.Limit = len(app.commits.Commits)
	app.historyScroll.SetOffset(0, len(app.commits.Commits), app.historyVisibleRows())
	app.refresh()

	clickID(t, app, "act-changes")

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw changes: %v", err)
	}

	view := app.View()
	if !view.History.Scroll {
		t.Fatalf("history does not scroll: commits=%d visible=%d", len(app.commits.Commits), app.historyVisibleRows())
	}

	if len(view.History.Commits) != app.historyVisibleRows() {
		t.Fatalf("visible commits = %d, want %d", len(view.History.Commits), app.historyVisibleRows())
	}

	if view.History.Commits[0].Lanes != 1 {
		t.Fatalf("linear parents produced %d lanes, want 1", view.History.Commits[0].Lanes)
	}

	boxes := app.Page().Boxes()
	if _, ok := boxByID(boxes, "history-scroll"); !ok {
		t.Fatal("no history scroll rail")
	}

	if _, ok := boxByID(boxes, "history-thumb"); !ok {
		t.Fatal("no history scroll thumb")
	}

	for _, id := range []string{"history-prev", "history-next"} {
		if _, ok := boxByID(boxes, id); ok {
			t.Fatalf("history has a pagination control %q", id)
		}
	}

	pane, ok := boxByID(boxes, "history-pane")
	if !ok {
		t.Fatal("no history pane")
	}

	// A wheel notch over the pane moves three rows.
	before := app.historyScroll.Offset
	handled, err := app.onWheel(ctx, pane.X+pane.W/2, pane.Y+pane.H/2, 0, -1)
	if err != nil {
		t.Fatalf("wheel over history: %v", err)
	}

	if !handled {
		t.Fatal("wheel over the history pane was not handled")
	}

	if app.historyScroll.Offset <= before {
		t.Fatalf("wheel offset = %d, want more than %d", app.historyScroll.Offset, before)
	}

	thumb, ok := boxByID(app.Page().Boxes(), "history-thumb")
	if !ok {
		t.Fatal("no history thumb after the wheel")
	}

	claimed, err := app.onDragStart(ctx, thumb)
	if err != nil {
		t.Fatalf("thumb drag start: %v", err)
	}

	if !claimed {
		t.Fatal("history thumb did not claim the drag")
	}

	before = app.historyScroll.Offset
	if err := app.onDragMove(ctx, 0, 50); err != nil {
		t.Fatalf("thumb drag: %v", err)
	}

	maxOffset := len(app.commits.Commits) - app.historyVisibleRows()
	if app.historyScroll.Offset == before {
		t.Fatalf("thumb drag left the offset at %d", before)
	}

	if app.historyScroll.Offset < 0 || app.historyScroll.Offset > maxOffset {
		t.Fatalf("dragged offset = %d, want 0..%d", app.historyScroll.Offset, maxOffset)
	}

	// A drag past either end pins the window to the first or the last row.
	if err := app.onDragMove(ctx, 0, 1e6); err != nil {
		t.Fatalf("thumb drag to the end: %v", err)
	}

	if app.historyScroll.Offset != maxOffset {
		t.Fatalf("offset at the end = %d, want %d", app.historyScroll.Offset, maxOffset)
	}

	if err := app.onDragMove(ctx, 0, -1e6); err != nil {
		t.Fatalf("thumb drag to the top: %v", err)
	}

	if app.historyScroll.Offset != 0 {
		t.Fatalf("offset at the top = %d, want 0", app.historyScroll.Offset)
	}

	if err := app.onDragMove(ctx, 0, 1e6); err != nil {
		t.Fatalf("thumb drag back to the end: %v", err)
	}

	view = app.View()
	if !view.History.AtEnd {
		t.Fatalf("history not at end: offset=%d total=%d", app.historyScroll.Offset, len(app.commits.Commits))
	}

	last := view.History.Commits[len(view.History.Commits)-1]
	if want := app.commits.Commits[len(app.commits.Commits)-1].Hash; last.Hash != want {
		t.Fatalf("last visible commit = %q, want %q", last.Hash, want)
	}

	if want := app.commits.Commits[maxOffset].Hash; view.History.Commits[0].Hash != want {
		t.Fatalf("first visible commit = %q, want %q", view.History.Commits[0].Hash, want)
	}

	lastBox, ok := boxByID(app.Page().Boxes(), "commit-"+last.Short)
	if !ok {
		t.Fatal("last commit has no row box")
	}

	pane, ok = boxByID(app.Page().Boxes(), "history-pane")
	if !ok {
		t.Fatal("no history pane after the drag")
	}

	if lastBox.Y < pane.Y || lastBox.Y+lastBox.H > pane.Y+pane.H+1 {
		t.Fatalf("last row at %.0f..%.0f is outside the pane %.0f..%.0f",
			lastBox.Y, lastBox.Y+lastBox.H, pane.Y, pane.Y+pane.H)
	}

	if slack := pane.Y + pane.H - (lastBox.Y + lastBox.H); slack > lastBox.H {
		t.Fatalf("last row is %.0f px above the pane bottom, want it at the bottom", slack)
	}
}

// historyScrollCommits builds a newest-first linear history. Each commit names
// the next one as its parent, so the graph holds one lane.
func historyScrollCommits(count int) commitLog {
	commits := make([]gitCommit, count)

	for i := range commits {
		// The first seven hex digits are the short hash, so the counter leads
		// the full hash instead of padding it with zeros.
		hash := fmt.Sprintf("%07x%033x", i+1, i+1)
		commits[i] = gitCommit{
			Hash:    hash,
			Short:   hash[:7],
			Subject: fmt.Sprintf("Commit %02d", i+1),
			Author:  "Test Author",
			Date:    "2026-10-27",
		}

		if i+1 < count {
			commits[i].Parents = []string{fmt.Sprintf("%07x%033x", i+2, i+2)}
		}
	}

	return commitLog{
		Branch:  "feat/history-scroll",
		Head:    commits[0].Hash,
		Commits: commits,
		Note:    "Synthetic linear history for the History scroll test.",
		Sample:  true,
	}
}
