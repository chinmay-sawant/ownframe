package window

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestDevToolsToggleSwallowsF12 checks the Phase 3 exit: F12 flips the
// overlay and the page's key handler never sees the press or the release.
func TestDevToolsToggleSwallowsF12(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	swallow, err := s.devKeyStep(ebiten.KeyF12, true, modifiers{})
	if err != nil || !swallow {
		t.Fatalf("f12 press swallow = %v, err = %v", swallow, err)
	}

	if d.on {
		t.Fatal("f12 did not close the overlay")
	}
	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}

	if d.hovers != 1 {
		t.Fatalf("close hovers = %d, want 1", d.hovers)
	}

	if down, up := s.pageKeyStep(ebiten.KeyF12, true); down || up {
		t.Fatal("the page saw the f12 press")
	}

	swallow, err = s.devKeyStep(ebiten.KeyF12, false, modifiers{})
	if err != nil || !swallow {
		t.Fatalf("f12 release swallow = %v, err = %v", swallow, err)
	}

	if down, up := s.pageKeyStep(ebiten.KeyF12, false); down || up {
		t.Fatal("the page saw the f12 release")
	}
}

// TestDevToolsToggleIgnoresRepeatPulse checks a server repeat reported as a
// release and a press in the next frame does not toggle twice.
func TestDevToolsToggleIgnoresRepeatPulse(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	if _, err := s.devKeyStep(ebiten.KeyF12, true, modifiers{}); err != nil {
		t.Fatal(err)
	}

	if d.on {
		t.Fatal("the first press did not toggle")
	}

	if _, err := s.devKeyStep(ebiten.KeyF12, false, modifiers{}); err != nil {
		t.Fatal(err)
	}

	if _, err := s.devKeyStep(ebiten.KeyF12, true, modifiers{}); err != nil {
		t.Fatal(err)
	}

	if d.on {
		t.Fatal("the repeat pulse toggled again")
	}
}
