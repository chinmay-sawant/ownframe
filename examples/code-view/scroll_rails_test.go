package main

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestScrollRailsPaint(t *testing.T) {
	if html := buildHTML(); !strings.Contains(html, ".list-scroll") || !strings.Contains(html, ".hist-graph") {
		t.Fatal("scrollrail stylesheet is not in the built page")
	}

	app := newTestApp(t)
	app.mode = "changes"
	app.changes = changeSet{Branch: "test", Note: "test"}

	for i := 0; i < 30; i++ {
		app.changes.Files = append(app.changes.Files, changedFile{Path: fmt.Sprintf("file-%02d.go", i), Status: "M"})
	}

	app.statuses = statusMap(app.changes)
	app.selectFirstDiff()

	app.commits.Commits = nil
	for i := 0; i < 30; i++ {
		app.commits.Commits = append(app.commits.Commits, gitCommit{Hash: fmt.Sprintf("%040d", i), Short: fmt.Sprintf("%07d", i)})
	}

	app.changesScroll.SetOffset(0, len(app.changes.Files), app.changeVisibleRows())
	app.historyScroll.SetOffset(0, len(app.commits.Commits), app.historyVisibleRows())
	app.refresh()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if _, ok := boxByID(app.Page().Boxes(), "changes-scroll"); !ok {
		t.Fatalf("no changes scroll rail: scroll=%v files=%d visible=%d", app.View().Diff.Scroll, len(app.changes.Files), app.changeVisibleRows())
	}

	if _, ok := boxByID(app.Page().Boxes(), "changes-thumb"); !ok {
		t.Fatal("no changes scroll thumb")
	}

	if _, ok := boxByID(app.Page().Boxes(), "history-scroll"); !ok {
		t.Fatalf("no history scroll rail: scroll=%v commits=%d visible=%d", app.View().History.Scroll, len(app.commits.Commits), app.historyVisibleRows())
	}

	if _, ok := boxByID(app.Page().Boxes(), "history-thumb"); !ok {
		t.Fatal("no history scroll thumb")
	}
}
