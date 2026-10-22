package ui

import (
	"context"
	"testing"
)

// TestProgressDirtyStaysOnTheRow checks a byte update at 1908x999 dirties
// one queue line. A union down to the footer would repaint the history.
func TestProgressDirtyStaysOnTheRow(t *testing.T) {
	app, back := newTestApp(t)
	app.page.SetSize(1908, 999)
	ctx := context.Background()

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	active := benchRows(3)
	back.send(Update{Kind: UpdateActive, Active: active})
	back.send(Update{Kind: UpdateHistory, Page: &PageResponse{
		Gen: app.pager.gen, Rows: benchRows(50), Total: 100,
	}})

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	app.page.TakeDirty()
	row := active[0]
	row.Done = row.Total / 2
	back.send(Update{Kind: UpdateProgress, Row: &row})

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	rect, ok := app.page.TakeDirty()
	if !ok {
		t.Fatal("progress tick left no dirty rect")
	}

	if rect.Dy() > 80 {
		t.Fatalf("dirty height = %d, want one row", rect.Dy())
	}
}
