package motion

import (
	"testing"
	"time"
)

// TestRuntimeFinishedPlayerParksItsDamage pins the settled-frame rule: a
// player that did no work holds the pose it already presented, so only its
// first resting sample marks damage. A repeated rest pose marks nothing, and
// a rebind repaints from base once.
func TestRuntimeFinishedPlayerParksItsDamage(t *testing.T) {
	r := New()
	r.now = testClock
	d := animatedDisplay("running", false)

	if err := r.Bind(d); err != nil {
		t.Fatalf("bind: %v", err)
	}

	// A mid-run sample moves and marks.
	mid := r.Step(d, testClock().Add(50*time.Millisecond))
	if mid.Empty() {
		t.Fatal("mid-run step marked no damage")
	}
	if got := r.Stats().Active; got != 1 {
		t.Fatalf("mid-run active = %d, want 1", got)
	}

	// The finishing sample still marks: the end pose is new paint.
	end := r.Step(d, testClock().Add(150*time.Millisecond))
	if end.Empty() {
		t.Fatal("finishing step marked no damage")
	}
	if got := r.Stats().Active; got != 0 {
		t.Fatalf("finished active = %d, want 0", got)
	}

	// Repeated resting samples paint the same pose and mark nothing.
	for i := range 3 {
		rest := r.Step(d, testClock().Add(time.Duration(200+i*50)*time.Millisecond))
		if !rest.Empty() {
			t.Fatalf("rest step %d marked %v, want no damage", i, rest.Bounds())
		}
	}

	// A rebind repaints from the declared base once, then parks again.
	rebound := animatedDisplay("running", false)
	if err := r.Bind(rebound); err != nil {
		t.Fatalf("rebind: %v", err)
	}
	first := r.Step(rebound, testClock().Add(300*time.Millisecond))
	if first.Empty() {
		t.Fatal("first step after rebind marked no damage")
	}
	second := r.Step(rebound, testClock().Add(350*time.Millisecond))
	if !second.Empty() {
		t.Fatalf("second step after rebind marked %v, want no damage", second.Bounds())
	}
}
