package main

import (
	"math"
	"strings"

	"github.com/chinmay-sawant/ownframe"
)

type rowPaintCache struct {
	generation uint64
	rows       map[string]*ownframe.DisplayOp
}

func paintPosition(x, y float64) [2]int64 {
	return [2]int64{int64(math.Round(x * coordScale)), int64(math.Round(y * coordScale))}
}

// coordScale fixes op coordinates to integers for the paint-position keys.
const coordScale = 1000

func (c *rowPaintCache) bind(p *ownframe.Page) {
	if c.generation == p.Probe().Generation() {
		return
	}
	c.generation = p.Probe().Generation()
	c.rows = map[string]*ownframe.DisplayOp{}
	d := p.Probe().Display()
	positions := map[[2]int64]string{}
	for _, b := range p.Boxes() {
		if strings.HasPrefix(b.ID, "row-") {
			positions[paintPosition(b.X*d.PointsPerPixel, b.Y*d.PointsPerPixel)] = b.ID
		}
	}
	for i := range d.Ops {
		op := &d.Ops[i]
		if op.Kind != ownframe.DisplayOpFillRect {
			continue
		}
		if id := positions[paintPosition(op.X, op.Y)]; id != "" {
			c.rows[id] = op
		}
	}
}
