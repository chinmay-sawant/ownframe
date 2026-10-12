package dump

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// bounds is the bounding box of rects, or the zero rect when the list is
// empty. Frame.Rect keeps this shape so a reader that only wants one region
// still sees the area the old one-rect page would have painted.
func bounds(rects []image.Rectangle) Rect {
	r := image.Rectangle{}
	for _, q := range rects {
		r = r.Union(q)
	}
	return Rect{X0: r.Min.X, Y0: r.Min.Y, X1: r.Max.X, Y1: r.Max.Y}
}

// areaOf sums the damage-list areas. TakeDirtyRects pads each rect out
// before it returns, so the sum can spend a few CSS pixels more than the
// marked boxes.
func areaOf(rects []image.Rectangle) int {
	n := 0
	for _, r := range rects {
		if !r.Empty() {
			n += r.Dx() * r.Dy()
		}
	}
	return n
}

// meetsAny counts the ops whose raw X, Y, W, H box meets any rect, once
// each. It approximates the real replay filter (internal/replay opTouches),
// which converts point-space bounds to CSS pixels and pads them for strokes
// and antialiasing; this test compares the raw fields, so ReplayedOps is an
// estimate, wanted only for the dump.
func meetsAny(d *layout.Display, rects []image.Rectangle) int {
	n := 0
	for i := range d.Ops {
		op := &d.Ops[i]
		box := image.Rect(int(op.X), int(op.Y), int(op.X+op.W), int(op.Y+op.H))
		for _, r := range rects {
			if !box.Intersect(r).Empty() {
				n++
				break
			}
		}
	}
	return n
}
