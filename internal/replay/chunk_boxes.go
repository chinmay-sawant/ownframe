package replay

import (
	"image"
	"sort"

	"github.com/chinmay-sawant/blinkless/layout"
)

// Chunk caps: a run holds at most this many ops, and its baked box at most
// this many pixels a side and in total. Past a cap the run splits; an op
// that cannot fit alone stays direct.
const (
	chunkMaxOps  = 64
	chunkMaxSide = 512
	chunkMaxArea = 1 << 20
	chunkTol     = 2
)

// chunkAreaHalf halves the content area to reject page-sized layers.
const chunkAreaHalf = 2

// chunkBox is one element box a chunk may own.
type chunkBox struct {
	index int
	box   image.Rectangle
}

// displayFingerprint is the plan cache generation for one display: the shape
// a redraw replaces wholesale. Lengths, canvas size, and scale distinguish a
// fresh layout from the previous one; the same pointer with the same shape
// reuses the cached group and chunk plans, while a splice that grows Order
// or Boxes changes the fingerprint and rebuilds. Tick patches mutate op
// values in place without changing the shape, so runs keep a separate
// op-write generation (see chunk_sig.go) to invalidate cached signatures.
func displayFingerprint(d *layout.Display) uint64 {
	if d == nil {
		return 0
	}

	h := fnvOffsetBasis
	h = mix64(h, uint64(len(d.Ops)))
	h = mix64(h, uint64(len(d.Order)))
	h = mix64(h, uint64(len(d.Boxes)))
	h = mix64(h, uint64(len(d.OpOwners)))
	h = mix64(h, uint64(d.Width))
	h = mix64(h, uint64(d.Height))
	h = mixF(h, d.PointsPerPixel)

	return h
}

// chunkBoxes returns the element boxes a chunk may use, smallest first. A
// box larger than half the content is a page-sized layer, not a card, so it
// never owns a chunk.
func chunkBoxes(d *layout.Display) []chunkBox {
	content := contentRect(d)
	half := content.Dx() * content.Dy() / chunkAreaHalf
	out := make([]chunkBox, 0, len(d.Boxes))

	for i := range d.Boxes {
		b := contentBox(d.Boxes[i])
		if area := b.Dx() * b.Dy(); area > 0 && area <= half {
			out = append(out, chunkBox{index: i, box: b})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].box.Dx()*out[i].box.Dy() < out[j].box.Dx()*out[j].box.Dy()
	})

	return out
}
