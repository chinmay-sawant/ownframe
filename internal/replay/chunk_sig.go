package replay

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
)

// chunkSig returns the content signature and the page box of one run. An
// unbounded member or a box past the caps gives an empty box, so the caller
// keeps the run direct.
func chunkSig(d *layout.Display, c *chunkRun, content image.Rectangle) (uint64, image.Rectangle) {
	h := uint64(14695981039346656037)
	box := image.Rectangle{}

	for _, op := range c.ops {
		b, ok := opBounds(op, d.PointsPerPixel)
		if !ok || b.Empty() {
			return 0, image.Rectangle{}
		}

		box = box.Union(b)
		h = sigOp(h, op)
	}

	box = box.Intersect(content)
	if box.Empty() || !chunkFits(box) {
		return 0, image.Rectangle{}
	}

	return h, box
}

// mix64 folds one word into an FNV-1a style signature.
func mix64(h, v uint64) uint64 {
	return (h ^ v) * 1099511628211
}

// mixF folds one float.
func mixF(h uint64, v float64) uint64 {
	return mix64(h, math.Float64bits(v))
}

// mixStr folds a string.
func mixStr(h uint64, s string) uint64 {
	for i := 0; i < len(s); i++ {
		h = mix64(h, uint64(s[i]))
	}

	return h
}

// bit packs one bool into a word.
func bit(v bool) uint64 {
	if v {
		return 1
	}

	return 0
}
