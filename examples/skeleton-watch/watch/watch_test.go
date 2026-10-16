package watch

import (
	"context"
	"testing"
	"time"
)

func TestWatchInit(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if app.Page() == nil {
		t.Fatal("Page() returned nil")
	}
	if app.Reserve().Hours() != 48.0 {
		t.Fatalf("expected 48h reserve, got %f", app.Reserve().Hours())
	}
	if app.Lume() {
		t.Fatal("expected lume to start disabled")
	}
	if app.Chrono().Running() {
		t.Fatal("expected chrono to start stopped")
	}
}

func TestTimeAngles(t *testing.T) {
	refTime := time.Date(2026, 10, 17, 3, 0, 0, 0, time.UTC)
	h, m, s := timeAngles(refTime)
	if h != 90.0 {
		t.Fatalf("expected hour angle 90.0, got %f", h)
	}
	if m != 0.0 || s != 0.0 {
		t.Fatalf("expected 0 for min and sec, got %f, %f", m, s)
	}
}

func TestTourbillonEscapement(t *testing.T) {
	t0 := time.Date(2026, 10, 17, 0, 0, 15, 0, time.UTC)
	cage, _ := tourbillonAngles(t0, false)
	if cage != 90.0 {
		t.Fatalf("expected tourbillon cage 90.0, got %f", cage)
	}

	cageStopped, balStopped := tourbillonAngles(t0, true)
	if cageStopped != 0 || balStopped != 0 {
		t.Fatalf("expected stopped escapement 0, got %f, %f", cageStopped, balStopped)
	}
}

func TestLumeToggle(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	app.ToggleLume()
	if !app.Lume() || app.view.LumeClass != "lume-on" {
		t.Fatalf("expected lume-on, got %v / %s", app.Lume(), app.view.LumeClass)
	}
	app.ToggleLume()
	if app.Lume() || app.view.LumeClass != "lume-off" {
		t.Fatalf("expected lume-off, got %v / %s", app.Lume(), app.view.LumeClass)
	}
}

func TestRedraw(t *testing.T) {
	app, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if err := app.Redraw(context.Background()); err != nil {
		t.Fatalf("Redraw() error: %v", err)
	}
}
