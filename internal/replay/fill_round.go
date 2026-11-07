package replay

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/chinmay-sawant/blinkless/layout"
)

var vectorPathPool = sync.Pool{
	New: func() any { return new(vector.Path) },
}

var roundedFillPathPool = &vectorPathPool

// fillRounded draws a fill with circular corners as one vector path. The
// engine's painter keeps the real corners when the canvas clips the box, so
// the path is built from the op rectangle itself.
func fillRounded(dst *ebiten.Image, op *layout.DisplayOp, radii [4]float64, dx, dy float64) {
	x := float32(op.X*pxPerPt + dx)
	y := float32(op.Y*pxPerPt + dy)
	w := float32(op.W * pxPerPt)
	h := float32(op.H * pxPerPt)
	r := radius4(radii)

	path := vectorPathPool.Get().(*vector.Path)
	path.Reset()
	populateRoundedFillPath(path, x, y, w, h, r)

	var options vector.DrawPathOptions
	options.AntiAlias = true
	clr := rgba(op)
	cr, cg, cb, ca := clr.RGBA()
	options.ColorScale.Scale(float32(cr)/0xffff, float32(cg)/0xffff, float32(cb)/0xffff, float32(ca)/0xffff)

	vector.FillPath(dst, path, nil, &options)
	path.Reset()
	vectorPathPool.Put(path)
}

// populateRoundedFillPath traces a fill's corners as circular arcs from its origin into path.
func populateRoundedFillPath(path *vector.Path, x, y, w, h float32, r [4]float32) {
	path.MoveTo(x+r[0], y)
	path.LineTo(x+w-r[1], y)
	arcTo(path, x+w, y, x+w, y+r[1], r[1])
	path.LineTo(x+w, y+h-r[2])
	arcTo(path, x+w, y+h, x+w-r[2], y+h, r[2])
	path.LineTo(x+r[3], y+h)
	arcTo(path, x, y+h, x, y+h-r[3], r[3])
	path.LineTo(x, y+r[0])
	arcTo(path, x, y, x+r[0], y, r[0])
	path.Close()
}

// roundedFillPath traces a fill's corners as circular arcs from its origin.
func roundedFillPath(x, y, w, h float32, r [4]float32) vector.Path {
	var path vector.Path
	populateRoundedFillPath(&path, x, y, w, h, r)
	return path
}

// radius4 scales a corner radius set from layout points to canvas pixels.
func radius4(radii [4]float64) [4]float32 {
	return [4]float32{
		float32(radii[0] * pxPerPt), float32(radii[1] * pxPerPt),
		float32(radii[2] * pxPerPt), float32(radii[3] * pxPerPt),
	}
}

func arcTo(path *vector.Path, x1, y1, x2, y2, radius float32) {
	if radius > 0 {
		path.ArcTo(x1, y1, x2, y2, radius)
	}
}
