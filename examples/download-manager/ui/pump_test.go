package ui

import (
	"context"
	"testing"
	"time"
)

func TestPumpRetriesDroppedRequests(t *testing.T) {
	app, back := newTestApp(t)
	ctx := context.Background()

	back.setBusy(true)
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if !app.needActive {
		t.Fatal("needActive cleared after a failed Active request")
	}

	if !app.historyDirty {
		t.Fatal("historyDirty cleared after a failed Page request")
	}

	back.setBusy(false)
	app.lastHistory = time.Time{}
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if app.needActive {
		t.Fatal("needActive not cleared after the retry")
	}

	if app.historyDirty {
		t.Fatal("historyDirty not cleared after the retry")
	}

	if len(back.pages) == 0 || back.actives == 0 {
		t.Fatalf("retries not recorded: pages=%d actives=%d", len(back.pages), back.actives)
	}
}

func TestPaintRebindsAfterRedraw(t *testing.T) {
	app, back := newTestApp(t)
	ctx := context.Background()

	back.send(Update{Kind: UpdateActive, Active: []Row{running("j1", 10, 100)}})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if app.bindings.fill[0] == nil {
		t.Fatal("fill not bound")
	}

	before := app.bindings.fill[0]

	// Advance the download so the redraw replaces the display list with new
	// progress. An identical redraw keeps the list and the binding.
	row := running("j1", 50, 100)
	back.send(Update{Kind: UpdateProgress, Row: &row})
	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if app.bindings.fill[0] == before {
		t.Fatal("the operation was not reacquired after the redraw")
	}
}

func TestPaintWithoutDisplay(t *testing.T) {
	app, _ := newTestApp(t)

	// New has not drawn yet, so there is no display list. A tick must not
	// panic on the bitmap path.
	app.paint()
}
