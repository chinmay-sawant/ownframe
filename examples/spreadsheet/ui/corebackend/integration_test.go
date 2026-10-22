package corebackend

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/spreadsheet/ui"
)

// waitFor ticks the screen until cond holds or the deadline passes.
func waitFor(t *testing.T, app *ui.App, cond func() bool) {
	t.Helper()

	ctx := context.Background()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if err := app.Page().Testing().Tick(ctx); err != nil {
			t.Fatal(err)
		}

		if cond() {
			return
		}

		time.Sleep(2 * time.Millisecond)
	}

	t.Fatal("condition not reached")
}

// hasCell reports whether the view prints one cell with the text.
func hasCell(app *ui.App, text string) bool {
	for _, c := range app.View().Cells {
		if c.Text == text {
			return true
		}
	}

	return false
}

func TestUIShowsSeededWorkbook(t *testing.T) {
	b := newBackend(t, t.TempDir())
	app, err := ui.New(ui.Options{Backend: b, Width: 1000, Height: 700, DataDir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = app.Close() }()

	waitFor(t, app, func() bool { return hasCell(app, "row 1") })

	if got := app.View().TotalH; got != ui.ChromeH+workbookMaxRows*ui.RowH {
		t.Fatalf("TotalH = %d", got)
	}
}

const workbookMaxRows = 1048576
