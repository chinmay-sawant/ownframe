package watch

import "github.com/chinmay-sawant/ownframe"

func (a *App) bind(d *ownframe.Display) {
	boxes := a.page.Boxes()
	p := &a.parts

	const dialCenterX = 315.0
	const dialCenterY = 315.0

	p.handHour = bindPart(d, boxes, "hand-hour", dialCenterX, dialCenterY)
	p.handMinute = bindPart(d, boxes, "hand-minute", dialCenterX, dialCenterY)
	p.handSecond = bindPart(d, boxes, "hand-second", dialCenterX, dialCenterY)

	p.tourbillon = bindPart(d, boxes, "tourbillon-cage", 315.0, 419.25)
	p.chronoMin = bindPart(d, boxes, "hand-chrono-min", 315.0, 227.25)
	p.powerHand = bindPart(d, boxes, "hand-power", 217.5, 324.75)

	a.bound = a.page.Generation()
}

func bindPart(d *ownframe.Display, boxes []ownframe.Box, id string, px, py float64) partAssembly {
	const pts = 0.75
	for _, b := range boxes {
		if b.ID != id {
			continue
		}
		bx, by, bw, bh := b.X*pts, b.Y*pts, b.W*pts, b.H*pts
		var ops []*ownframe.DisplayOp
		var bases []mat6
		for i := range d.Ops {
			op := &d.Ops[i]
			if op.X >= bx-2.0 && op.X+op.W <= bx+bw+2.0 &&
				op.Y >= by-2.0 && op.Y+op.H <= by+bh+2.0 {
				ops = append(ops, op)
				bases = append(bases, baseOf(op))
			}
		}
		return partAssembly{ops: ops, base: bases, px: px, py: py}
	}
	return partAssembly{}
}
