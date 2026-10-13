package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// chunkBoxFor finds the smallest element box that contains the op's core
// layout box within the tolerance, and returns the op's painted box too. An
// op that is fixed, z-indexed, translucent, unbounded, or too large to chunk
// reports false. A translucent op keeps its run direct: baking it over a
// transparent image and compositing that image can round a channel one step
// away from a direct draw, so parity would not be byte-exact.
func chunkBoxFor(d *layout.Display, op *layout.DisplayOp, p *chunkPlan) (int, image.Rectangle, bool) {
	if op.Fixed || op.ZIndexSet || op.Opacity() < 1 {
		return -1, image.Rectangle{}, false
	}

	b, ok := opBounds(op, d.PointsPerPixel)
	if !ok || b.Empty() || !chunkFits(b) {
		return -1, image.Rectangle{}, false
	}

	core, ok := chunkCoreBox(op, d.PointsPerPixel)
	if !ok {
		core = b
	}

	best := p.grid.find(p.content, p.boxes, core)
	if best < 0 {
		return -1, image.Rectangle{}, false
	}

	return p.boxes[best].index, b, true
}

// chunkFits reports a box the caps allow.
func chunkFits(b image.Rectangle) bool {
	return b.Dx() <= chunkMaxSide && b.Dy() <= chunkMaxSide && b.Dx()*b.Dy() <= chunkMaxArea
}
