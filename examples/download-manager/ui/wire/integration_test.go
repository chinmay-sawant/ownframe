package wire

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/ui"
)

// TestEndToEndDummy runs the example headless against the real store,
// scheduler, and fake transport: seeded data, a page of history, a manual
// add, and the persisted theme.
func TestEndToEndDummy(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	backend := newBackend(t, Config{DataDir: dir, Dummy: true})

	app, err := ui.New(ctx, backend)
	if err != nil {
		_ = backend.Close()
		t.Fatalf("ui.New: %v", err)
	}

	defer func() {
		if err := app.Close(); err != nil {
			t.Errorf("Close: %v", err)
		}
	}()

	tick(t, app, 30)

	view := app.View()
	if view.Summary.Completed < 1 {
		t.Fatalf("summary = %+v", view.Summary)
	}

	if len(view.Active) == 0 {
		t.Fatal("no seeded active jobs")
	}

	if len(view.History) == 0 {
		t.Fatal("no history rows on the first page")
	}

	if len(view.History) > ui.HistoryPage {
		t.Fatalf("history page has %d rows", len(view.History))
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	app.Page().SetFormValue("url", "https://example.invalid/manual.bin")
	clickVisible(t, app, "add")
	waitSeen(t, app, "https://example.invalid/manual.bin", 3*time.Second)

	clickVisible(t, app, "theme")
	waitFile(t, filepath.Join(dir, "ui.json"), 3*time.Second)

	if backend.Close() != nil {
		t.Fatal("the second Close did not return nil")
	}
}
