package watch

import (
	"context"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestActionsAndShortcuts(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	ctx := context.Background()

	// Pusher click starts chrono.
	if err := app.onClick(ctx, ownframe.Box{ID: "pusher-start"}); err != nil {
		t.Fatalf("click pusher-start: %v", err)
	}
	if !app.Chrono().Running() {
		t.Fatal("expected chrono running after start click")
	}

	// Space key toggles chrono to stop.
	if err := app.onKeyDown(ctx, "space"); err != nil {
		t.Fatalf("key space: %v", err)
	}
	if app.Chrono().Running() {
		t.Fatal("expected chrono stopped after space key")
	}

	// Crown click winds reserve.
	initHours := app.Reserve().Hours()
	if err := app.onClick(ctx, ownframe.Box{ID: "crown"}); err != nil {
		t.Fatalf("click crown: %v", err)
	}
	if app.Reserve().Hours() <= initHours {
		t.Fatal("expected reserve to increase after crown click")
	}

	// L key toggles lume.
	wasLume := app.Lume()
	if err := app.onKeyDown(ctx, "l"); err != nil {
		t.Fatalf("key l: %v", err)
	}
	if app.Lume() == wasLume {
		t.Fatal("expected lume to toggle on 'l' key")
	}
}

func TestTickUpdates(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	ctx := context.Background()
	fixed := time.Date(2026, 10, 17, 10, 8, 42, 0, time.UTC)
	app.SetNow(func() time.Time { return fixed })

	if err := app.Tick(ctx); err != nil {
		t.Fatalf("Tick() error: %v", err)
	}
	if app.view.ClockText != "10:08:42" {
		t.Fatalf("expected clock 10:08:42, got %s", app.view.ClockText)
	}

	// Deplete reserve and ensure movement stops.
	app.Reserve().Drain(72 * 3600)
	if err := app.Tick(ctx); err != nil {
		t.Fatalf("Tick() error: %v", err)
	}
	if app.view.Status != "STOPPED (Depleted)" {
		t.Fatalf("expected STOPPED status, got %s", app.view.Status)
	}
}
