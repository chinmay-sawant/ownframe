package scene

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

func TestStaleSettingsAreDropped(t *testing.T) {
	m := &fakeModel{}
	st := &fakeStore{}
	s, clock := newTestScene(t, m, st, Options{Stepper: &fakeStepper{}})

	st.push(Result{ID: st.settingsID + 99, Kind: ResultSettings})
	tickAt(t, s, clock, time.Second/60)

	if len(m.configs) != 0 {
		t.Fatal("a stale settings result was applied")
	}
}

func TestSettingsApplyThemeAndRestore(t *testing.T) {
	m := &fakeModel{}
	st := &fakeStore{}
	s, clock := newTestScene(t, m, st, Options{Stepper: &fakeStepper{}})

	st.push(Result{ID: st.settingsID, Kind: ResultSettings, Set: SettingsResult{
		Settings: Settings{Dark: true},
		Snapshot: game.Snapshot{ID: "g1"},
		Has:      true,
	}})

	gen := s.page.Generation()
	tickAt(t, s, clock, time.Second/60)

	if !s.dark {
		t.Fatal("the saved dark theme was not applied")
	}

	if s.page.Generation() == gen {
		t.Fatal("the theme change did not redraw")
	}

	if len(m.restores) != 1 {
		t.Fatalf("restores = %d, want 1", len(m.restores))
	}

	if len(m.configs) != 1 {
		t.Fatalf("configures = %d, want 1", len(m.configs))
	}
}
