package corebackend

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/ui"
)

func TestUITypeCommitUndoOverStore(t *testing.T) {
	dir := t.TempDir()
	b := newBackend(t, dir)
	app, err := ui.New(ui.Options{Backend: b, Width: 1000, Height: 700, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	page := app.Page()
	waitFor(t, app, func() bool { return hasCell(app, "row 1") })

	// Replace A1 and commit with Enter. The status clears when the worker
	// acknowledges the save.
	if err := page.Testing().Type(ctx, "edited"); err != nil {
		t.Fatal(err)
	}

	if err := page.Testing().Submit(ctx); err != nil {
		t.Fatal(err)
	}

	waitFor(t, app, func() bool { return hasCell(app, "edited") && app.View().Status == "" })

	// Undo through the page chord, then wait for the acknowledgement.
	if err := page.Undo(ctx); err != nil {
		t.Fatal(err)
	}

	waitFor(t, app, func() bool { return hasCell(app, "row 1") && app.View().Status == "undo" })

	// Stop the screen, then read the database through a fresh backend.
	if err := app.Close(); err != nil {
		t.Fatal(err)
	}

	again := newBackend(t, dir)
	id := sheetID(t, again, "Numbers")
	cells, err := again.Range(id, ui.Area{R0: 0, C0: 0, R1: 0, C1: 0})
	if err != nil {
		t.Fatal(err)
	}

	if cells[0].Raw != "row 1" {
		t.Fatalf("stored cell after undo = %+v", cells[0])
	}
}

func TestUICtrlEndJumpsToUsedRange(t *testing.T) {
	b := newBackend(t, t.TempDir())
	app, err := ui.New(ui.Options{Backend: b, Width: 1000, Height: 700, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = app.Close() }()

	ctx := context.Background()
	waitFor(t, app, func() bool { return hasCell(app, "row 1") })

	if err := app.Page().Testing().KeyDown(ctx, "ctrl+end"); err != nil {
		t.Fatal(err)
	}

	waitFor(t, app, func() bool { return app.View().Ref == "T200" })
}
