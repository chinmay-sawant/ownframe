package wire

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/download-manager/fixture"
	"github.com/chinmay-sawant/ownframe/examples/download-manager/ui"
)

// TestRestartKeepsData completes a real fixture transfer, reopens the same
// data directory, and checks that the completed row and the theme survived.
func TestRestartKeepsData(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	first := newBackend(t, Config{DataDir: dir, Fixture: true})
	app, err := ui.New(ctx, first)
	if err != nil {
		_ = first.Close()
		t.Fatalf("ui.New: %v", err)
	}

	if err := app.Redraw(ctx); err != nil {
		t.Fatalf("Redraw: %v", err)
	}

	url := first.fixture.URL + fixture.PathOK
	app.Page().SetFormValue("url", url)
	clickVisible(t, app, "add")
	waitHistory(t, app, url, ui.StateCompleted, 15*time.Second)

	clickVisible(t, app, "theme")

	if err := app.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}

	second := newBackend(t, Config{DataDir: dir})
	app2, err := ui.New(ctx, second)
	if err != nil {
		_ = second.Close()
		t.Fatalf("second ui.New: %v", err)
	}

	defer func() {
		if err := app2.Close(); err != nil {
			t.Errorf("second Close: %v", err)
		}
	}()

	if dark, err := second.Dark(); err != nil || !dark {
		t.Fatalf("theme after restart: dark=%v err=%v", dark, err)
	}

	tick(t, app2, 30)

	if app2.View().Summary.Completed < 1 {
		t.Fatalf("summary after restart = %+v", app2.View().Summary)
	}

	found := false
	for _, row := range app2.View().History {
		if row.URL == url && row.State == ui.StateCompleted {
			found = true
		}
	}

	if !found {
		t.Fatal("the completed job did not survive the restart")
	}
}
