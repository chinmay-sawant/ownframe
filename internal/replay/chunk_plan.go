package replay

import (
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// chunkPaint is one step of a display's chunk plan: a direct op, one cached
// run, or a non-normal group node painted as a unit.
type chunkPaint struct {
	op    *layout.DisplayOp
	run   *chunkRun
	group *groupNode
}

// chunkRun is a consecutive run of ops that share one element box.
type chunkRun struct {
	id  int
	ops []*layout.DisplayOp
}

// chunkPlan is a display's paint order with eligible runs folded into chunks.
// boxes are the element boxes a chunk may use, smallest first, and grid
// indexes them so an op only tests the boxes near it.
type chunkPlan struct {
	items   []chunkPaint
	content image.Rectangle
	boxes   []chunkBox
	grid    *chunkGrid
}

// buildChunks flattens the group tree into a paint plan and folds runs of
// ops that share an element box into chunks. A non-normal or masked group
// stays a unit, so its subtree keeps direct ops and the group painters never
// meet a chunk.
func buildChunks(d *layout.Display, root *groupNode) *chunkPlan {
	p := &chunkPlan{content: contentRect(d), boxes: chunkBoxes(d)}
	p.grid = newChunkGrid(p.content, p.boxes)
	next := 0
	flattenNode(d, root, p, &next)

	return p
}
