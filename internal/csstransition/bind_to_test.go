package csstransition

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/cssmotion"
)

// TestBindToAnimatesToTheCapturedDestination checks that a rebind with an
// explicit destination animates toward the captured paint. The operation
// carries a mid-transition value, so the destination is the captured paint
// and not the value the display holds.
func TestBindToAnimatesToTheCapturedDestination(t *testing.T) {
	display, box, indices, from, to := colorCaptures(t)
	if len(indices) != 1 || len(from) != 1 || len(to) != 1 {
		t.Fatalf("paint counts = %d/%d/%d", len(indices), len(from), len(to))
	}

	ops := cssmotion.Resolve(display, indices)
	cssmotion.Write(ops[0], cssmotion.Blend(from[0], to[0], .5))

	start := time.Unix(5, 0)
	r, err := BindTo(display, box, "background-color", indices, from, to, start)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Step(start.Add(500 * time.Millisecond)) {
		t.Fatal("the transition stopped at its midpoint")
	}

	mid := cssmotion.Read(ops[0])
	if mid.R <= 0 || mid.R >= 1 {
		t.Fatalf("midpoint paint = %+v, want an interpolated channel", mid)
	}

	r.Step(start.Add(time.Second))
	if got := cssmotion.Read(ops[0]); got != to[0] {
		t.Fatalf("final paint = %+v, want the captured destination %+v", got, to[0])
	}
}

// TestBindToRefusesAMismatchedDestination checks that a destination paint
// count that does not match the operations is refused.
func TestBindToRefusesAMismatchedDestination(t *testing.T) {
	display, box, indices, from, _ := colorCaptures(t)
	final := append(append([]cssmotion.Paint{}, from...), from...)
	if _, err := BindTo(display, box, "background-color", indices, from, final, time.Unix(6, 0)); err == nil {
		t.Fatal("a destination paint count that does not match was accepted")
	}
}
