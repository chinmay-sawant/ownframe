package replay

import (
	"image"
	"sync/atomic"

	"github.com/chinmay-sawant/blinkless/layout"
)

// chunkPaint is one step of a display's chunk plan: a direct op, one cached
// run, or a non-normal group node painted as a unit.
type chunkPaint struct {
	op    *layout.DisplayOp
	run   *chunkRun
	group *groupNode
}

// chunkRun is a consecutive run of ops that share one element box. sig and
// box are computed once at plan build; writeGen is the op-write generation
// they were built from, so a repaint reuses them without rehashing while the
// tick-writable fields still match.
type chunkRun struct {
	id       int
	ops      []*layout.DisplayOp
	sig      uint64
	box      image.Rectangle
	writeGen uint64
	hasSig   bool
}

// chunkPlan is a display's paint order with eligible runs folded into chunks.
// boxes are the element boxes a chunk may use, smallest first, and grid
// indexes them so an op only tests the boxes near it. tiles buckets each
// item by 256 px tile once at build, so a tile repaint looks up candidates
// instead of scanning the whole plan per tile. gen is the display shape the
// plan was built from.
type chunkPlan struct {
	items   []chunkPaint
	content image.Rectangle
	boxes   []chunkBox
	grid    *chunkGrid
	tiles   *planTileIndex
	gen     uint64
}

// chunkPlanBuilds counts chunk-plan folds. Tests prove reuse by observing no
// build across identical repaints, not by timing.
var chunkPlanBuilds atomic.Int64 //nolint:gochecknoglobals // reason: one process-wide counter read by chunkPlanBuildCount so tests can prove plan reuse.

// chunkPlanBuildCount reports chunk-plan folds.
func chunkPlanBuildCount() int64 {
	return chunkPlanBuilds.Load()
}

// buildChunks flattens the group tree into a paint plan and folds runs of
// ops that share an element box into chunks. A non-normal or masked group
// stays a unit, so its subtree keeps direct ops and the group painters never
// meet a chunk. Signatures compute once here; tile buckets index every item
// once here.
func buildChunks(d *layout.Display, root *groupNode) *chunkPlan {
	p := &chunkPlan{content: contentRect(d), boxes: chunkBoxes(d), gen: displayFingerprint(d)}
	p.grid = newChunkGrid(p.content, p.boxes)
	next := 0
	flattenNode(d, root, p, &next)
	sealChunkSigs(d, p)
	indexPlanTiles(d, p)
	chunkPlanBuilds.Add(1)

	return p
}

// sealChunkSigs computes each run's signature and box once at plan build and
// records the write generation. A run that cannot bake keeps hasSig false.
func sealChunkSigs(d *layout.Display, p *chunkPlan) {
	for i := range p.items {
		c := p.items[i].run
		if c == nil {
			continue
		}

		sig, box := chunkSig(d, c, p.content)
		if box.Empty() {
			continue
		}

		c.sig = sig
		c.box = box
		c.writeGen = runWriteGen(c)
		c.hasSig = true
	}
}

// indexPlanTiles buckets every plan item by tile once at build. Groups use
// their page box; unbounded ink becomes a global visited by every tile.
func indexPlanTiles(d *layout.Display, p *chunkPlan) {
	t := &planTileIndex{origin: p.content.Min}
	t.cols = planTileSpan(p.content.Dx())
	t.rows = planTileSpan(p.content.Dy())
	t.buckets = make([][]int, t.cols*t.rows)

	ppt := 0.0
	if d != nil {
		ppt = d.PointsPerPixel
	}

	for i := range p.items {
		it := &p.items[i]

		switch {
		case it.run != nil:
			if it.run.hasSig {
				t.add(it.run.box, i, true)
			} else {
				t.add(image.Rectangle{}, i, false)
			}
		case it.op != nil:
			box, ok := opBounds(it.op, ppt)
			t.add(box, i, ok)
		case it.group != nil:
			box, ok := nodeBox(it.group, ppt)
			t.add(box, i, ok)
		}
	}

	p.tiles = t
}
