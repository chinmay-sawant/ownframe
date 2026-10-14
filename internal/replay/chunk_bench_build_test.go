package replay

import (
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
)

const (
	benchCards = 20
	benchCardW = 200
	benchCardH = 90
	benchGap   = 10
	benchCols  = 4
)

// cardBenchDisplay builds a grid of opaque cards, five ops each: a
// background, a rounded header, a border, a divider, and a chip. Every op
// sits in the card's element box, so one card folds into one run. Boxes are
// CSS pixels and op coordinates are points, as layout emits them.
func cardBenchDisplay() *layout.Display {
	const ppt = 72.0 / 96.0
	pt := func(css float64) float64 { return css * ppt }

	rows := benchCards / benchCols
	d := &layout.Display{
		Width:          benchCols*benchCardW + (benchCols+1)*benchGap,
		Height:         rows*benchCardH + (rows+1)*benchGap,
		PointsPerPixel: ppt,
	}

	for i := range benchCards {
		x := float64(benchGap + (i%benchCols)*(benchCardW+benchGap))
		y := float64(benchGap + (i/benchCols)*(benchCardH+benchGap))
		d.Boxes = append(d.Boxes, layout.Box{
			ID: fmt.Sprintf("card%d", i), X: x, Y: y, W: benchCardW, H: benchCardH,
		})

		base := len(d.Ops)
		d.Ops = append(d.Ops,
			layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: pt(x + 2), Y: pt(y + 2),
				W: pt(benchCardW - 4), H: pt(benchCardH - 4), R: .9, G: .92, B: .96},
			layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: pt(x + 2), Y: pt(y + 2),
				W: pt(benchCardW - 4), H: pt(22), Radius: pt(4), R: .2, G: .4, B: .8},
			layout.DisplayOp{Kind: layout.DisplayOpStrokeRect, X: pt(x + 2), Y: pt(y + 2),
				W: pt(benchCardW - 4), H: pt(benchCardH - 4), Width: pt(2), R: .2, G: .25, B: .3},
			layout.DisplayOp{Kind: layout.DisplayOpLine, X: pt(x + 6), Y: pt(y + 40),
				W: pt(benchCardW - 12), Width: pt(1), R: .7, G: .7, B: .75},
			layout.DisplayOp{Kind: layout.DisplayOpFillRect, X: pt(x + 8), Y: pt(y + 58),
				W: pt(40), H: pt(16), R: .1, G: .6, B: .3},
		)

		for j := base; j < len(d.Ops); j++ {
			d.Order = append(d.Order, j)
		}
	}

	return d
}
