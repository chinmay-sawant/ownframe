package page

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

// opBoundsXform maps an op's local painted box through its baked 2D
// transform, so a partial repaint covers where the ink actually lands.
func opBoundsXform(op *layout.DisplayOp, d *layout.Display) image.Rectangle {
	scale := d.PixelPerPoint
	if scale <= 0 {
		scale = 1
	}

	x0, y0, x1, y1 := localPaintBox(op)
	m := op.Transform()

	loX, loY := math.Inf(1), math.Inf(1)
	hiX, hiY := math.Inf(-1), math.Inf(-1)

	for _, p := range [4][2]float64{{x0, y0}, {x1, y0}, {x1, y1}, {x0, y1}} {
		x, y := m.Apply(p[0], p[1])
		loX, loY = math.Min(loX, x), math.Min(loY, y)
		hiX, hiY = math.Max(hiX, x), math.Max(hiY, y)
	}

	return image.Rect(
		int(loX*scale)-1, int(loY*scale)-1,
		int(hiX*scale)+1, int(hiY*scale)+1,
	)
}

// localPaintBox is the op's painted box in points before its transform,
// matching the geometry the replay bounds use.
func localPaintBox(op *layout.DisplayOp) (x0, y0, x1, y1 float64) {
	x0, y0 = op.X, op.Y
	x1, y1 = op.X+op.W, op.Y+op.H

	if op.Kind == layout.DisplayOpText || op.Kind == layout.DisplayOpBullet {
		y0 = op.Y - ascentOf(op)
		y1 = op.Y + op.InkDescent
	}

	if op.Kind == layout.DisplayOpStrokeRect || op.Kind == layout.DisplayOpLine {
		grow := math.Max(op.Width, 1) / 2
		x0, y0, x1, y1 = x0-grow, y0-grow, x1+grow, y1+grow
	}

	return x0, y0, x1, y1
}
