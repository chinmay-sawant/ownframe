package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"math"
	"sort"
	"sync"
	"sync/atomic"
)

// get returns the entry for key, baking it when the slot is missing or its
// signature or box moved. A changed signature with an unchanged box reuses
// the image and paints over it. The bake stays under the lock, as measured
// by BenchmarkChunkCardDamage/chunked: bake draws into a slot image a
// concurrent bake could share, and Ebiten calls belong on the draw path,
// so moving it out races. The replay suite stays green under -race with
// the bake inline. Fewer misses, not a split lock, cuts stall time.
func (c *chunkStore) get(key chunkKey, sig uint64, box image.Rectangle,
	bake func(*ebiten.Image)) *chunkEntry {
	c.mu.Lock()
	defer c.mu.Unlock()

	old := c.entries[key]
	if old != nil && old.sig == sig && old.box == box {
		c.touch(key)
		return old
	}

	if c.entries == nil {
		c.entries = make(map[chunkKey]*chunkEntry)
	}

	var img *ebiten.Image
	if old != nil && old.box == box {
		old.img.Clear()
		img = old.img
	} else {
		img = ebiten.NewImage(box.Dx(), box.Dy())
		if old != nil {
			c.pixels -= old.img.Bounds().Dx() * old.img.Bounds().Dy()
			old.img.Deallocate()
		}

		c.pixels += box.Dx() * box.Dy()
	}

	bake(img)

	if old == nil {
		c.order = append(c.order, key)
	} else {
		c.touch(key)
	}

	e := &chunkEntry{img: img, sig: sig, box: box}
	c.entries[key] = e
	c.evict(key)

	return e
}

// chunkMaxEntries and chunkMaxPixels bound the chunk images one renderer
// keeps: at most 128 chunks holding 2M pixels (8 MB of GPU bytes).
// Least-recently used evicts first, and its image is disposed. A paint
// only needs the chunks of the region it draws, so the cap covers the
// visible rows rather than the whole page.
const (
	chunkMaxEntries = 128
	chunkMaxPixels  = 2 << 20
)

// chunkKey names one chunk's cache slot: the sequence of the run in a
// display's chunk plan. A display is rebuilt on every redraw, so the key
// carries no display identity: the content signature decides whether the
// slot still holds the same pixels, so a redraw with the same content reuses
// the baked image instead of baking it again, and the cache no longer pins
// the displays it came from.
//
// The signature has to survive a redraw for that to pay off, so it folds
// values a layout reproduces (see chunk_sig_op.go) and never a per-run
// pointer. A renderer owns these slots, so a second page cannot replace them.
type chunkKey struct {
	id int
}

// chunkEntry is one baked chunk: the image, the content signature it holds,
// and the page box its origin maps to.
type chunkEntry struct {
	img *ebiten.Image
	sig uint64
	box image.Rectangle
}

type chunkStore struct {
	mu      sync.Mutex
	entries map[chunkKey]*chunkEntry
	order   []chunkKey
	pixels  int
}

// replayCounts count the work a paint actually did: single ops drawn and
// chunk images blitted. A benchmark turns the gate on; shipped frames pay
// one atomic load per draw.
var replayCounts struct {
	on    atomic.Bool
	ops   atomic.Int64
	blits atomic.Int64
}

func countOp() {
	if replayCounts.on.Load() {
		replayCounts.ops.Add(1)
	}
}

func countBlit() {
	if replayCounts.on.Load() {
		replayCounts.blits.Add(1)
	}
}

// touch marks key most-recently used by moving it to the back of the order,
// so a later evict drops the least-recently used slot first. Every get path
// calls it: a hit, a signature rebuild, and a box rebuild all mean the chunk
// painted this frame, and the next paint only needs the chunks of the region
// it draws.
func (c *chunkStore) touch(key chunkKey) {
	for i, candidate := range c.order {
		if candidate != key {
			continue
		}
		c.order = append(c.order[:i], c.order[i+1:]...)
		c.order = append(c.order, key)
		return
	}
}

// evict drops the oldest slots while the store is over either bound. protected
// is the entry just inserted or rebuilt and is never evicted here, so an entry
// that is itself larger than the pixel budget cannot be disposed and then
// returned to its caller.
func (c *chunkStore) evict(protected chunkKey) {
	for len(c.entries) > chunkMaxEntries || c.pixels > chunkMaxPixels {
		victim := chunkKey{}
		idx := -1
		for i, candidate := range c.order {
			if candidate == protected {
				continue
			}
			victim = candidate
			idx = i
			break
		}
		if idx < 0 {
			// Only the protected entry remains; let it stay.
			break
		}

		c.order = append(c.order[:idx], c.order[idx+1:]...)

		if e := c.entries[victim]; e != nil {
			delete(c.entries, victim)
			c.pixels -= e.img.Bounds().Dx() * e.img.Bounds().Dy()
			e.img.Deallocate()
		}
	}
}

// chunkCoreBox is the op's layout box in CSS pixels, before stroke growth:
// the geometry the element box should contain. The painted box can reach
// half a border width outside it, so assignment tests this box and the bake
// still covers opBounds.
func chunkCoreBox(op *layout.DisplayOp, ppt float64) (image.Rectangle, bool) {
	if op.XformSet && !op.Transform().IsIdentity() {
		return xformBounds(op, ppt)
	}

	switch op.Kind {
	case layout.DisplayOpText, layout.DisplayOpBullet:
		return textBounds(op, ppt), true
	case layout.DisplayOpLine:
		x, y, w, h, _ := op.PaintLineGeometry()

		return boxBounds(min(x, x+w), min(y, y+h), math.Abs(w), math.Abs(h), 0, ppt), true
	case layout.DisplayOpGridRun:
		return gridBounds(op, ppt), true
	case layout.DisplayOpFillRect, layout.DisplayOpImage, layout.DisplayOpStrokeRect:
		return boxBounds(op.X, op.Y, op.W, op.H, 0, ppt), true
	}

	return image.Rectangle{}, false
}

// chunkBoxContains reports whether box holds inner within the tolerance.
func chunkBoxContains(box, inner image.Rectangle) bool {
	return inner.Min.X >= box.Min.X-chunkTol && inner.Min.Y >= box.Min.Y-chunkTol &&
		inner.Max.X <= box.Max.X+chunkTol && inner.Max.Y <= box.Max.Y+chunkTol
}

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
	h = mix64(h, uint64(d.Width))  //nolint:gosec // reason: display width is a non-negative layout size, kept in the hash as its bit pattern.
	h = mix64(h, uint64(d.Height)) //nolint:gosec // reason: display height is a non-negative layout size, kept in the hash as its bit pattern.
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

// chunkPlanFor returns the display's chunk plan from the same two-entry
// cache that holds the group tree. nil means the entry was evicted while it
// was built; the caller then keeps the direct walk.
func (renderer *Renderer) chunkPlanFor(d *layout.Display) *chunkPlan {
	root := renderer.planFor(d)

	renderer.plans.mu.Lock()
	defer renderer.plans.mu.Unlock()

	for i := range renderer.plans.entries {
		if renderer.plans.entries[i].d == d {
			if renderer.plans.entries[i].plan == nil {
				renderer.plans.entries[i].plan = buildChunks(d, root)
			}

			return renderer.plans.entries[i].plan
		}
	}

	return nil
}

// chunkRaster returns the baked image for one run, or nil when the run
// cannot make one: an unbounded op, an empty box, or a box past the caps.
func (renderer *Renderer) chunkRaster(d *layout.Display, p *chunkPlan, c *chunkRun) *chunkEntry {
	sig, box := chunkSig(d, c, p.content)
	if box.Empty() {
		return nil
	}

	key := chunkKey{id: c.id}

	return renderer.chunks.get(key, sig, box, func(img *ebiten.Image) {
		for _, op := range c.ops {
			countOp()
			renderer.drawOp(img, op, -float64(box.Min.X), -float64(box.Min.Y))
		}
	})
}
