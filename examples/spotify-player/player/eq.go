package player

import (
	"math"
	"slices"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

// eqHeight is bar i's height as a fraction of the container, 0.2 to 1.
func eqHeight(i int, playing bool, t time.Time) float64 {
	if !playing {
		return 0.3
	}

	f := 0.5 + 0.5*math.Sin(float64(t.UnixMilli())/float64(120+35*i)+float64(i)*1.9)
	if f < 0.2 {
		return 0.2
	}

	if f > 1 {
		return 1
	}

	return f
}

// eqBars is reused across frames so animateEq does not allocate a fresh
// slice and backing array on every tick.
var eqBars []*ownframe.DisplayOp

// animateEq bottom-aligns the accent bars and pumps them while playing.
func (a *App) animateEq(d *ownframe.Display, boxes []ownframe.Box) {
	eq, ok := boxByID(boxes, "eq")
	if !ok {
		return
	}

	if d == nil {
		return
	}

	x, y, w, h := frame.BoxUnits(d, eq)

	eqBars = eqBars[:0]

	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != ownframe.DisplayOpFillRect {
			continue
		}

		cx, cy := op.X+op.W/2, op.Y+op.H/2
		if !(cx >= x-0.5 && cx <= x+w+0.5 && cy >= y-0.5 && cy <= y+h+0.5) {
			continue
		}

		const tolerance = 0.02
		if !(op.Alpha > 0.5 &&
			math.Abs(op.R-accent[0]) < tolerance &&
			math.Abs(op.G-accent[1]) < tolerance &&
			math.Abs(op.B-accent[2]) < tolerance) {
			continue
		}

		eqBars = append(eqBars, op)
	}

	if len(eqBars) == 0 {
		return
	}

	slices.SortFunc(eqBars, func(a, b *ownframe.DisplayOp) int {
		switch {
		case a.X < b.X:
			return -1
		case a.X > b.X:
			return 1
		default:
			return 0
		}
	})

	playing := a.audio.Playing()
	now := time.Now()

	for i, op := range eqBars {
		barH := eqHeight(i, playing, now) * h
		op.H = barH
		op.Y = y + h - barH
	}

	a.page.Invalidate("eq")
}
