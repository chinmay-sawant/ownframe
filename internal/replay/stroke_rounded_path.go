package replay

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2/vector"
)

// roundedPathAt traces a border centerline with per-corner elliptical arcs
// from its origin.
func roundedPathAt(x, y, w, h float32, rx, ry [4]float32) vector.Path {
	var path vector.Path

	path.MoveTo(x+rx[0], y)
	path.LineTo(x+w-rx[1], y)
	ellipseTo(&path, x+w-rx[1], y+ry[1], rx[1], ry[1], -math.Pi/2, 0)
	path.LineTo(x+w, y+h-ry[2])
	ellipseTo(&path, x+w-rx[2], y+h-ry[2], rx[2], ry[2], 0, math.Pi/2)
	path.LineTo(x+rx[3], y+h)
	ellipseTo(&path, x+rx[3], y+h-ry[3], rx[3], ry[3], math.Pi/2, math.Pi)
	path.LineTo(x, y+ry[0])
	ellipseTo(&path, x+rx[0], y+ry[0], rx[0], ry[0], math.Pi, 3*math.Pi/2)
	path.Close()

	return path
}

// ellipseTo appends an elliptical arc as short segments. The path must
// already sit on the arc's start point.
func ellipseTo(path *vector.Path, cx, cy, rx, ry, from, to float32) {
	const steps = 8

	for i := 1; i <= steps; i++ {
		a := from + (to-from)*float32(i)/steps
		path.LineTo(cx+rx*float32(math.Cos(float64(a))), cy+ry*float32(math.Sin(float64(a))))
	}
}
