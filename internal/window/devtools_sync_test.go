package window

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// No window or GPU loop runs here: a pointer action must leave the drawn
// viewport unchanged until the next sync, which precedes resize.
func TestDevToolsToggleDefersGeometryUntilSync(t *testing.T) {
	d := newDevScreen()
	s := newDevShell(d)
	beforeW, beforeH := s.viewSize()
	if err := s.devToggle(); err != nil {
		t.Fatal(err)
	}
	if d.DevTools() {
		t.Fatal("close request did not reach the page")
	}
	if w, h := s.viewSize(); w != beforeW || h != beforeH || d.hovers != 0 {
		t.Fatal("pointer close changed geometry before the next resize pass")
	}
	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}
	if s.dev.on || d.hovers != 1 {
		t.Fatal("sync did not close the dock and clear hover exactly once")
	}
	if w, h := s.viewSize(); w != s.screenW || h != s.screenH {
		t.Fatalf("closed viewport = %dx%d", w, h)
	}
}

// TestDevToolsKeepsF12AvailableAfterPageClose guards the walk hook case: the
// page opens the overlay, then closes it before the first frame. F12 must
// still reopen the dock.
func TestDevToolsKeepsF12AvailableAfterPageClose(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.available, s.dev.on = false, false

	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}
	if !s.dev.on || !s.dev.available {
		t.Fatal("page opt-in did not make the overlay available")
	}

	d.on = false
	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}
	if s.dev.on || !s.dev.available {
		t.Fatal("page close dropped F12 availability")
	}

	swallow, err := s.devKeyStep(ebiten.KeyF12, true, modifiers{})
	if err != nil || !swallow {
		t.Fatalf("f12 after page close: swallow=%v err=%v", swallow, err)
	}
	if !d.on {
		t.Fatal("f12 did not reopen the overlay")
	}
}
