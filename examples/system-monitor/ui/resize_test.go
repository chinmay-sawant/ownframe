package ui

import (
	"context"
	"testing"
)

func TestTickRebindsAfterResize(t *testing.T) {
	app := newTestApp(t, &fakeSource{}, nil)
	ctx := context.Background()

	app.mail.putSummary(app.currentGen(), sample("cpu", 0.9))

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	app.page.SetSize(920, 620)

	if err := app.page.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := app.Tick(ctx); err != nil {
		t.Fatal(err)
	}

	if app.state.h.gen != app.page.Generation() {
		t.Fatal("handles not reacquired after a resize relayout")
	}

	if w, h := app.page.Size(); w != 920 || h != 620 {
		t.Fatalf("size = %dx%d", w, h)
	}

	bars := app.state.h.bars["cpu"]
	if len(bars) == 0 || maxBar(bars) <= 0 {
		t.Fatal("bars not repainted after resize")
	}
}

func TestClampKeepsMinimumWindow(t *testing.T) {
	app := newTestApp(t, &fakeSource{}, nil)

	app.page.SetSize(200, 100)

	if w, h := app.page.Size(); w != minWidth || h != minHeight {
		t.Fatalf("clamped size = %dx%d, want %dx%d", w, h, minWidth, minHeight)
	}
}
