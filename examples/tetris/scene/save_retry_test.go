package scene

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
)

func TestFailedSaveKeepsRetrying(t *testing.T) {
	m := &fakeModel{}
	st := &fakeStore{}
	s, clock := newTestScene(t, m, st, Options{Stepper: &fakeStepper{}})

	res := game.Result{ID: "g3", Score: 2}
	m.done = &res

	tickAt(t, s, clock, time.Second/60)

	st.push(Result{ID: st.saveID, Kind: ResultSaved, Err: errTest})
	tickAt(t, s, clock, time.Second/60)

	if s.pendingSave == nil {
		t.Fatal("a failed save was dropped")
	}

	if s.status != "SCORE NOT SAVED - RETRYING" {
		t.Fatalf("status = %q", s.status)
	}
}

func TestCloseSavesSnapshotAndJoins(t *testing.T) {
	m := &fakeModel{}
	st := &fakeStore{}
	s, _ := newTestScene(t, m, st, Options{Stepper: &fakeStepper{}})

	m.snap = game.Snapshot{ID: "resume"}
	m.snapOK = true

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	if len(st.snaps) != 1 {
		t.Fatalf("snapshots = %d, want 1", len(st.snaps))
	}

	if !st.closed {
		t.Fatal("the store was not closed")
	}

	if s.page.Probe().Ticking() {
		t.Fatal("the tick was not removed")
	}
}

func TestRestartClearsTheResumeSlot(t *testing.T) {
	m := &fakeModel{f: Frame{Phase: game.PhaseRunning}, run: "run-1"}
	st := &fakeStore{}
	s, clock := newTestScene(t, m, st, Options{Stepper: &fakeStepper{}})

	m.run = "run-2"
	tickAt(t, s, clock, time.Second/60)

	if st.clears != 1 {
		t.Fatalf("resume clears = %d, want 1", st.clears)
	}
}
