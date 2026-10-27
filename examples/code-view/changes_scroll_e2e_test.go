package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// TestChangesScrollE2E drives the Source Control changed-file list through a
// wheel and a thumb drag. The list is a vertical scroll: it has a rail and a
// thumb, not a pager, and both inputs move the same Offset to the clamped ends.
func TestChangesScrollE2E(t *testing.T) {
	ctx := context.Background()
	app := newTestApp(t)

	app.changes = changeSet{Branch: "feature/changes-scroll", Note: "Synthetic changed-file list."}
	for i := 0; i < 30; i++ {
		app.changes.Files = append(app.changes.Files, changedFile{
			Path:   fmt.Sprintf("changed/file-%02d.go", i),
			Status: "M",
			Add:    i + 1,
			Del:    i / 2,
		})
	}

	app.statuses = statusMap(app.changes)
	app.selectFirstDiff()

	if err := app.onClick(ctx, ownframe.Box{Action: "mode-changes"}); err != nil {
		t.Fatalf("switch to changes mode: %v", err)
	}

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw changes mode: %v", err)
	}

	view := app.View()
	if !view.Diff.Scroll {
		t.Fatalf("Diff.Scroll = false with %d files and %d visible rows", len(app.changes.Files), app.changeVisibleRows())
	}

	boxes := app.Page().Boxes()
	if _, ok := boxByID(boxes, "changes-scroll"); !ok {
		t.Fatal("changes-scroll box is missing")
	}

	thumb, ok := boxByID(boxes, "changes-thumb")
	if !ok {
		t.Fatal("changes-thumb box is missing")
	}

	for _, box := range boxes {
		if box.ID == "diff-page-next" || box.ID == "diff-page-prev" {
			t.Fatalf("changes list contains pager box %q", box.ID)
		}
	}

	list, ok := boxByID(boxes, "changes-list")
	if !ok {
		t.Fatal("changes-list box is missing")
	}

	beforeWheel := app.changesScroll.Offset
	handled, err := app.onWheel(ctx, list.X+list.W/2, list.Y+list.H/2, 0, -1)
	if err != nil {
		t.Fatalf("wheel over changes list: %v", err)
	}

	if !handled {
		t.Fatal("wheel over changes list was not handled")
	}

	if app.changesScroll.Offset <= beforeWheel {
		t.Fatalf("wheel offset = %d, want greater than %d", app.changesScroll.Offset, beforeWheel)
	}

	total := len(app.changes.Files)
	visible := app.changeVisibleRows()
	maxOffset := changesMaxOffset(total, visible)
	if maxOffset < 1 {
		t.Fatalf("max offset = %d for %d files and %d visible rows", maxOffset, total, visible)
	}

	beforeDrag := app.changesScroll.Offset
	claimed, err := app.onDragStart(ctx, thumb)
	if err != nil {
		t.Fatalf("claim changes-thumb drag: %v", err)
	}

	if !claimed {
		t.Fatal("changes-thumb drag was not claimed")
	}

	if err := app.onDragMove(ctx, 0, 40); err != nil {
		t.Fatalf("drag changes-thumb: %v", err)
	}

	if app.changesScroll.Offset == beforeDrag {
		t.Fatalf("thumb drag kept offset at %d", beforeDrag)
	}

	if err := app.onDragMove(ctx, 0, 100000); err != nil {
		t.Fatalf("drag changes-thumb past the end: %v", err)
	}

	if got := app.changesScroll.Offset; got != maxOffset {
		t.Fatalf("thumb drag past the end = %d, want clamped %d", got, maxOffset)
	}

	if err := app.onDragMove(ctx, 0, -100000); err != nil {
		t.Fatalf("drag changes-thumb past the start: %v", err)
	}

	if got := app.changesScroll.Offset; got != 0 {
		t.Fatalf("thumb drag past the start = %d, want clamped 0", got)
	}

	if err := app.onDragEnd(ctx); err != nil {
		t.Fatalf("end changes-thumb drag: %v", err)
	}

	for i := 0; i < total && !app.View().Diff.AtEnd; i++ {
		handled, err := app.onWheel(ctx, list.X+list.W/2, list.Y+list.H/2, 0, -1)
		if err != nil {
			t.Fatalf("wheel to bottom at step %d: %v", i, err)
		}

		if !handled {
			t.Fatalf("wheel to bottom at step %d was not handled", i)
		}
	}

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw bottom: %v", err)
	}

	view = app.View()
	if !view.Diff.AtEnd {
		t.Fatalf("Diff.AtEnd = false at offset %d of %d", app.changesScroll.Offset, maxOffset)
	}

	if len(view.Diff.Files) == 0 {
		t.Fatal("no visible changed files at the bottom")
	}

	last := app.changes.Files[len(app.changes.Files)-1].Path
	if got := view.Diff.Files[len(view.Diff.Files)-1].Path; got != last {
		t.Fatalf("last visible file = %q, want %q", got, last)
	}
}
