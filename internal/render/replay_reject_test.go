package render_test

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// TestReplayRejectsEllipticalRadius pins the shape the vector painter cannot
// draw: a fill with rx != ry. The gate must keep this page on the bitmap path.
func TestReplayRejectsEllipticalRadius(t *testing.T) {
	t.Parallel()

	display := acceptDisplay(t, `<div style="width:200px;height:80px;`+
		`background:#eee;border-radius:50% / 20%">oval</div>`)

	if render.Replayable(display) {
		t.Fatal("elliptical fill should not replay")
	}

	var elliptical bool

	for i := range display.Ops {
		if display.Ops[i].Kind != render.OpFillRect {
			continue
		}

		if _, ok := masking.FillRadii(&display.Ops[i]); !ok {
			elliptical = true
		}
	}

	if !elliptical {
		t.Fatal("no elliptical fill reached the gate; rejection is for another reason")
	}
}
