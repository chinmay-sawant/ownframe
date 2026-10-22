package game

import (
	"slices"
	"testing"
	"time"
)

func TestSnapshotRoundTrip(t *testing.T) {
	g := New(9)
	g.Start()
	g.Apply(ActionLeft)
	g.Apply(ActionSoftDrop)
	g.Step(1200 * time.Millisecond)

	snap := g.Snapshot()

	back, err := FromSnapshot(snap)
	if err != nil {
		t.Fatal(err)
	}

	expectPhase(t, back, PhasePaused)

	if back.ID != g.ID || back.Score != g.Score || back.Lines != g.Lines ||
		back.Level != g.Level || back.Pieces != g.Pieces || back.Elapsed != g.Elapsed {
		t.Fatalf("scalars differ: %+v vs %+v", back.Result(), g.Result())
	}

	if back.Piece != g.Piece || back.Rot != g.Rot || back.X != g.X || back.Y != g.Y {
		t.Fatal("the piece differs after the round trip")
	}

	if !slices.Equal(back.Board.rows(), g.Board.rows()) {
		t.Fatal("the board differs after the round trip")
	}

	if !slices.Equal(back.Next, g.Next) || !slices.Equal(back.bag, g.bag) {
		t.Fatal("the queues differ after the round trip")
	}

	if back.rng.state != g.rng.state {
		t.Fatal("the generator state differs after the round trip")
	}
}

func TestTwoResumesDivergeTogether(t *testing.T) {
	g := New(11)
	g.Start()

	snap := g.Snapshot()

	a, err := FromSnapshot(snap)
	if err != nil {
		t.Fatal(err)
	}

	b, err := FromSnapshot(snap)
	if err != nil {
		t.Fatal(err)
	}

	a.Start()
	b.Start()

	for i := 0; i < 60; i++ {
		a.Step(FixedStep)
		b.Step(FixedStep)
	}

	if a.Piece != b.Piece || a.X != b.X || a.Y != b.Y || !slices.Equal(a.Next, b.Next) {
		t.Fatal("two resumes of one snapshot diverged")
	}
}
