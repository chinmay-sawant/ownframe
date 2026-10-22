package scene

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

func TestPhaseChangeRedrawsOverlay(t *testing.T) {
	m := &fakeModel{f: Frame{Phase: game.PhaseRunning}}
	s, clock := newTestScene(t, m, nil, Options{Stepper: &fakeStepper{}})

	m.f.Phase = game.PhasePaused
	gen := s.page.Generation()

	tickAt(t, s, clock, time.Second/60)

	if s.page.Generation() == gen {
		t.Fatal("the phase change did not redraw")
	}

	if !hasText(s, "PAUSED") {
		t.Fatal("the paused overlay is missing")
	}

	if hasText(s, "PRESS ENTER") {
		t.Fatal("the ready overlay is still drawn")
	}
}
