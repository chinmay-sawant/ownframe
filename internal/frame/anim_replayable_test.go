package frame

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// A patched transform must stay on the replay path: Replayable is what keeps
// a page off the bitmap fallback, and scale and rotate add a baked matrix.
func TestPatchedXformStaysReplayable(t *testing.T) {
	t.Parallel()

	d, op := cardDisplay()
	if !render.Replayable(d) {
		t.Fatal("card display is not replayable before the animation")
	}

	anims := []*Anim{
		ScaleTo(d, []*page.DisplayOp{op}, 2, time.Millisecond, Linear),
		RotateTo(d, []*page.DisplayOp{op}, 30, time.Millisecond, Linear),
	}

	for _, a := range anims {
		a.Step(t0)
		a.Step(t0.Add(time.Millisecond))

		if !render.Replayable(d) {
			t.Fatalf("display not replayable after animation kind %v", a.kind)
		}
	}
}
