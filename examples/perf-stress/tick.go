package main

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// accent is the progress green, #1db954.
var accent = [3]float64{29.0 / 255, 185.0 / 255, 84.0 / 255}

// tick state: bound tracks the display generation, bars the eq fills.
type tickState struct {
	bound uint64
	seek  *ownframe.DisplayOp
	bars  []*ownframe.DisplayOp
	label *ownframe.DisplayOp
}

// Tick advances the counter and repaints progress and eq in place.
func (a *App) Tick(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	a.view.Tick++
	a.paint()
	return nil
}

// paint moves the seek fill, pumps the eq bars, and rewrites the labels.
func (a *App) paint() {
	d := a.page.Display()
	if d == nil {
		return
	}
	if a.page.Generation() != a.state.bound {
		a.bind(d)
	}
	frac := float64(a.view.Tick%progressCycle) / progressScale
	_, _, tw, _ := ownframe.BoxUnits(d, a.track)
	if a.state.seek != nil {
		a.state.seek.W = tw * frac
	}
	_, ey, _, eh := ownframe.BoxUnits(d, a.eq)
	now := float64(time.Now().UnixMilli())
	for i, op := range a.state.bars {
		f := eqMid + eqMid*math.Sin(now/float64(eqPeriodBase+eqPeriodStep*i)+float64(i)*eqPhaseStep)
		h := (eqFloor + eqSpan*f) * eh
		op.H = h
		op.Y = ey + eh - h
	}
	if a.state.label != nil {
		a.state.label.Text = strconv.Itoa(a.view.Tick)
	}
}
