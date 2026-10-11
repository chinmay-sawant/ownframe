package bitmap

import (
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

func shapeContains(op *layout.DisplayOp, x, y float64) bool {
	if op.Kind == layout.DisplayOpLine {
		lx, ly, w, h, width := op.PaintLineGeometry()
		return onLine(x, y, lx, ly, w, h, max(width, 1/pxPerPt))
	}
	rx, ry := render.RadiiXY(op)
	if op.Kind == layout.DisplayOpFillRect {
		return inRound(x-op.X, y-op.Y, op.W, op.H, rx, ry)
	}
	width := max(op.Width, 1/pxPerPt)
	if op.StrokeMask != 0 {
		for i, l := range [][4]float64{{op.X, op.Y, op.W, 0}, {op.X + op.W, op.Y, 0, op.H},
			{op.X, op.Y + op.H, op.W, 0}, {op.X, op.Y, 0, op.H}} {
			if op.StrokeMask&(1<<i) != 0 && onLine(x, y, l[0], l[1], l[2], l[3], width) {
				return true
			}
		}
		return false
	}
	half := width / 2
	outerX, outerY, innerX, innerY := rx, ry, rx, ry
	for i := range rx {
		outerX[i], outerY[i] = rx[i]+half, ry[i]+half
		innerX[i], innerY[i] = max(0, rx[i]-half), max(0, ry[i]-half)
	}
	return inRound(x-op.X+half, y-op.Y+half, op.W+width, op.H+width, outerX, outerY) &&
		!inRound(x-op.X-half, y-op.Y-half, op.W-width, op.H-width, innerX, innerY)
}

func onLine(x, y, lx, ly, w, h, width float64) bool {
	length := math.Hypot(w, h)
	if length == 0 {
		return false
	}
	along := ((x-lx)*w + (y-ly)*h) / length
	across := ((x-lx)*h - (y-ly)*w) / length
	return along >= -width/2 && along <= length+width/2 && math.Abs(across) <= width/2
}
