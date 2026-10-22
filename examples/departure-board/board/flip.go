package board

import (
	"image"
	"math/rand"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

const glyphs = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "

// flap is one tile's animation: the op it paints, its hit-test box, the glyph
// it is heading for, how many frames it waits, and how many it spins.
type flap struct {
	op          *ownframe.DisplayOp
	box         image.Rectangle
	want, cur   string
	delay, spin int
	done        bool
}

// flapRect is a hit-test box in whole CSS pixels, the same rounding the page
// uses, so MarkRect covers the tile.
func flapRect(b ownframe.Box) image.Rectangle {
	return image.Rect(int(b.X), int(b.Y), int(b.X+b.W+pixelRound), int(b.Y+b.H+pixelRound))
}

// flipper holds every tile of the board and the two header ops the tick
// moves. changed is scratch the step reuses so a frame allocates nothing, and
// barAt paces the refresh fill on wall-clock time.
type flipper struct {
	gen        uint64
	cells      map[string]*flap
	order      []string
	changed    []string
	rnd        *rand.Rand
	clock, bar *ownframe.DisplayOp
	clockText  string
	barFullW   float64
	barAt      time.Time
}

// step advances every flap one frame and returns the ids whose glyph changed
// this frame, reusing the changed slice. The caller must read the result
// before the next step.
func (f *flipper) step() []string {
	if f.rnd == nil {
		f.rnd = rand.New(rand.NewSource(1)) //nolint:gosec // reason: deterministic flap animation, not security.
	}

	f.changed = f.changed[:0]

	for _, id := range f.order {
		fl := f.cells[id]
		if fl == nil || fl.done {
			continue
		}

		if fl.delay > 0 {
			fl.delay--

			continue
		}

		if fl.spin > 0 {
			fl.spin--

			if ch := string(glyphs[f.rnd.Intn(len(glyphs))]); ch != fl.op.Text {
				fl.op.Text = ch
				f.changed = append(f.changed, id)
			}

			continue
		}

		if fl.op.Text != fl.want {
			fl.op.Text = fl.want
			f.changed = append(f.changed, id)
		}

		fl.done = true
	}

	return f.changed
}
