package main

import (
	"math"
	"strings"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

type rowPaintCache struct {
	generation uint64
	rows       map[string]*ownframe.DisplayOp
}

func paintPosition(x, y float64) [2]int64 {
	return [2]int64{int64(math.Round(x * 1000)), int64(math.Round(y * 1000))}
}

func (c *rowPaintCache) bind(p *page.Page) {
	if c.generation == p.Generation() {
		return
	}
	c.generation = p.Generation()
	c.rows = map[string]*ownframe.DisplayOp{}
	d := p.Display()
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
