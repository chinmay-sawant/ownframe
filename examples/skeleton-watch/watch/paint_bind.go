package watch

import (
	"math"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

func (a *App) bind(d *ownframe.Display) {
	boxes := a.page.Boxes()
	p := &a.parts
	p.handHour = fillAt(d, boxes, "hand-hour")
	p.handMinute = fillAt(d, boxes, "hand-minute")
	p.handSecond = fillAt(d, boxes, "hand-second")
	p.chronoSec = fillAt(d, boxes, "chrono-sec")
	p.chronoMin = fillAt(d, boxes, "chrono-min")
	p.tourbillon = fillAt(d, boxes, "tourbillon")
	p.balance = fillAt(d, boxes, "balance")
	p.powerHand = fillAt(d, boxes, "power-hand")
	p.gearBarrel = fillAt(d, boxes, "gear-barrel")
	p.gearCenter = fillAt(d, boxes, "gear-center")
	p.gearThird = fillAt(d, boxes, "gear-third")
	p.gearFourth = fillAt(d, boxes, "gear-fourth")

	p.textClock = textAt(d, boxes, "clock-display")
	p.textChrono = textAt(d, boxes, "chrono-display")
	p.textReserve = textAt(d, boxes, "reserve-display")
	p.textStatus = textAt(d, boxes, "status-display")

	a.bound = a.page.Generation()
}

func fillAt(d *ownframe.Display, boxes []ownframe.Box, id string) *ownframe.DisplayOp {
	for _, b := range boxes {
		if b.ID != id {
			continue
		}
		x, y, w, h := frame.BoxUnits(d, b)
		for i := range d.Ops {
			op := &d.Ops[i]
			if op.Kind != ownframe.DisplayOpFillRect {
				continue
			}
			if math.Abs(op.X-x) <= 2.0 && math.Abs(op.Y-y) <= 2.0 &&
				math.Abs(op.W-w) <= 2.0 && math.Abs(op.H-h) <= 2.0 {
				return op
			}
		}
	}
	return nil
}

func textAt(d *ownframe.Display, boxes []ownframe.Box, id string) *ownframe.DisplayOp {
	if d == nil {
		return nil
	}
	for _, b := range boxes {
		if b.ID != id {
			continue
		}
		x, y, w, h := frame.BoxUnits(d, b)
		for i := range d.Ops {
			op := &d.Ops[i]
			if op.Kind == ownframe.DisplayOpText &&
				op.X >= x-1.0 && op.X <= x+w+1.0 &&
				op.Y >= y-1.0 && op.Y <= y+h+1.0 {
				return op
			}
		}
	}
	return nil
}
