package dirty

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

// halfDivisor splits a stroke width into its two-sided growth.
const halfDivisor = 2

// maxDirtyScan caps the ops CountOps inspects, so one huge frame cannot
// make the dirty count cost more than the replay it measures.
const maxDirtyScan = 4096

// OpBounds is one operation's box in CSS pixels. A text or bullet op
// carries its baseline in Y, so the top comes from the face ascent. A
// transformed op bounds where its ink lands, not its layout box.
func OpBounds(op *layout.DisplayOp, d *layout.Display) image.Rectangle {
	if op.XformSet && !op.Transform().IsIdentity() {
		return opBoundsXform(op, d)
	}

	scale := d.PixelPerPoint
	if scale <= 0 {
		scale = 1
	}

	x0, y0 := op.X, op.Y
	x1, y1 := op.X+op.W, op.Y+op.H

	if op.Kind == layout.DisplayOpText || op.Kind == layout.DisplayOpBullet {
		y0 = op.Y - ascentOf(op)
		y1 = op.Y + op.InkDescent
	}

	return image.Rect(
		int(x0*scale)-1, int(y0*scale)-1,
		int(x1*scale)+1, int(y1*scale)+1,
	)
}

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
		grow := math.Max(op.Width, 1) / halfDivisor
		x0, y0, x1, y1 = x0-grow, y0-grow, x1+grow, y1+grow
	}

	return x0, y0, x1, y1
}

// ascentOf returns the face ascent in points for one text op.
func ascentOf(op *layout.DisplayOp) float64 {
	if op.Font == nil || op.Font.UnitsPerEm() <= 0 {
		return op.Size
	}

	return float64(op.Font.Ascent()) / float64(op.Font.UnitsPerEm()) * op.Size
}

// CountOps counts the ops whose bounds meet r, stopping at maxDirtyScan
// so the count stays cheaper than a full replay.
func CountOps(d *layout.Display, r image.Rectangle) int {
	if d == nil {
		return 0
	}

	n := 0

	for i := range d.Ops {
		if i >= maxDirtyScan {
			break
		}

		if OpBounds(&d.Ops[i], d).Intersect(r).Empty() {
			continue
		}

		n++
	}

	return n
}
