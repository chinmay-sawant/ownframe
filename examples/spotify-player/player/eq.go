package player

import (
	"math"
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

// animateEq bottom-aligns the accent bars and pumps them while playing.
func (a *App) animateEq(d *ownframe.Display, boxes []ownframe.Box) {
	eq, ok := boxByID(boxes, "eq")
	if !ok {
		return
	}

	bars := frame.Fills(d, eq, accent)
	if len(bars) == 0 {
		return
	}

	_, y, _, h := frame.BoxUnits(d, eq)
	playing := a.audio.Playing()
	now := time.Now()

	for i, op := range bars {
		barH := eqHeight(i, playing, now) * h
		op.H = barH
		op.Y = y + h - barH
	}

	a.page.Invalidate("eq")
}
