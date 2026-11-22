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

type lineCoverage struct {
	x, y, dx, dy float64
	length, half float64
}

func newLineCoverage(x, y, dx, dy, width float64) lineCoverage {
	return lineCoverage{
		x: x, y: y, dx: dx, dy: dy,
		length: math.Hypot(dx, dy),
		half:   width / halfDivisor,
	}
}

func (line *lineCoverage) contains(x, y float64) bool {
	x, y = x-line.x, y-line.y
	along := (x*line.dx + y*line.dy) / line.length
	across := (x*line.dy - y*line.dx) / line.length

	return along >= -line.half && along <= line.length+line.half && math.Abs(across) <= line.half
}

func maskedStrokeLines(op *layout.DisplayOp, width float64, lines []lineCoverage) []lineCoverage {
	sides := [4][4]float64{
		{op.X, op.Y, op.W, 0},
		{op.X + op.W, op.Y, 0, op.H},
		{op.X, op.Y + op.H, op.W, 0},
		{op.X, op.Y, 0, op.H},
	}
	for i, side := range sides {
		if op.StrokeMask&(1<<i) != 0 {
			line := newLineCoverage(side[0], side[1], side[2], side[3], width)
			if line.length != 0 {
				lines = append(lines, line)
			}
		}
	}

	return lines
}
