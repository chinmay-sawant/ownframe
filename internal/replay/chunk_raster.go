package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

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
