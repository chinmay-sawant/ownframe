package page

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// opBounds is one operation's box in CSS pixels. A text or bullet op carries
// its baseline in Y, so the top comes from the face ascent.
func opBounds(op *layout.DisplayOp, d *layout.Display) image.Rectangle {
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

// ascentOf returns the face ascent in points for one text op.
func ascentOf(op *layout.DisplayOp) float64 {
	if op.Font == nil || op.Font.UnitsPerEm() <= 0 {
		return op.Size
	}

	return float64(op.Font.Ascent()) / float64(op.Font.UnitsPerEm()) * op.Size
}
