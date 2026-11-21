package bitmap

import (
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

func strokeRadii(rx, ry [4]float64, half float64) (outerX, outerY, innerX, innerY [4]float64) {
	for i := range rx {
		outerX[i], outerY[i] = rx[i]+half, ry[i]+half
		innerX[i], innerY[i] = max(0, rx[i]-half), max(0, ry[i]-half)
	}
	return outerX, outerY, innerX, innerY
}

func onStrokeMask(op *layout.DisplayOp, x, y, width float64) bool {
	for i, l := range [4][4]float64{{op.X, op.Y, op.W, 0}, {op.X + op.W, op.Y, 0, op.H},
		{op.X, op.Y + op.H, op.W, 0}, {op.X, op.Y, 0, op.H}} {
		if op.StrokeMask&(1<<i) != 0 && onLine(x, y, l[0], l[1], l[2], l[3], width) {
			return true
		}
	}
	return false
}

func onLine(x, y, lx, ly, w, h, width float64) bool {
	length := math.Hypot(w, h)
	if length == 0 {
		return false
	}
	along := ((x-lx)*w + (y-ly)*h) / length
	across := ((x-lx)*h - (y-ly)*w) / length
	return along >= -width/halfDivisor && along <= length+width/halfDivisor && math.Abs(across) <= width/halfDivisor
}
