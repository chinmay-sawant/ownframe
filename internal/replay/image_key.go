package replay

import (
	"hash/fnv"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/blinkless/layout"
)

// imageBytesKey identifies one encoded payload with FNV-1a. The IsJPEG flag is
// folded in so identical bytes tagged two ways never share an entry.
func imageBytesKey(op *layout.DisplayOp) uint64 {
	if op == nil {
		return 0
	}

	data, _, _ := op.ImageBytes()

	hash := fnv.New64a()
	if op.IsJPEG {
		_, _ = hash.Write([]byte{1})
	}

	_, _ = hash.Write(data)

	return hash.Sum64()
}

// geomTransCol is the GeoM column holding translation.
const geomTransCol = 2

// imageGeom maps the decoded image's w by h pixels onto op's point rectangle,
// moved by the page offset and then through the op's CSS transform when it has
// one. The gate allows transformed images.
func imageGeom(op *layout.DisplayOp, dx, dy float64, w, h int) ebiten.GeoM {
	var geom ebiten.GeoM

	k := pxPerPt
	sx, sy := 1/k, 1/k
	if w > 0 {
		sx = op.W / float64(w)
	}
	if h > 0 {
		sy = op.H / float64(h)
	}

	a, b, c, d, e, f := 1.0, 0.0, 0.0, 1.0, 0.0, 0.0
	if op.XformSet {
		m := op.Transform()
		a, b, c, d, e, f = m.A, m.B, m.C, m.D, m.E, m.F
	}

	geom.SetElement(0, 0, k*a*sx)
	geom.SetElement(0, 1, k*c*sy)
	geom.SetElement(0, geomTransCol, k*(a*op.X+c*op.Y+e)+dx)
	geom.SetElement(1, 0, k*b*sx)
	geom.SetElement(1, 1, k*d*sy)
	geom.SetElement(1, geomTransCol, k*(b*op.X+d*op.Y+f)+dy)

	return geom
}
