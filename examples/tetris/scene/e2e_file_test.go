package scene

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/examples/tetris/game"
	"github.com/chinmay-sawant/ownframe/examples/tetris/store"
)

// TestEndToEndFileStore drives the whole scene against a fresh on-disk
// database, then reopens the same directory and reads the saved score.
func TestEndToEndFileStore(t *testing.T) {
	dir := t.TempDir()

	st, err := store.Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	step := &fakeStepper{}
	m := NewCore(21)
	s, clock := newTestScene(t, m, NewWorker(st), Options{Stepper: step})

	press(t, s, "enter")
	pump(t, s, clock, step, func() bool { return s.frame.Phase == game.PhaseRunning })

	press(t, s, "space")
	pump(t, s, clock, step, func() bool { return s.frame.Score > 0 })

	m.g.Phase = game.PhaseOver
	m.watch()
	pump(t, s, clock, step, func() bool { return s.status == "SCORE SAVED" })

	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	st2, err := store.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st2.Close() }()
	scores, err := st2.TopScores(context.Background(), 10)
	if err != nil {
		t.Fatal(err)
	}

	if len(scores) != 1 {
		t.Fatalf("scores = %d, want 1", len(scores))
	}

	if scores[0].ID != m.g.ID || scores[0].Score != m.g.Score {
		t.Fatalf("saved score = %+v", scores[0])
	}
}
