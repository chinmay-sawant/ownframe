package replay

import (
	"image"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// tilePlanFor returns the display's chunk plan, keyed on the display
// generation fingerprint. The group tree comes from the generation-keyed
// plan cache; the chunk fold builds lazily on the first tile paint that needs
// it, so a frame that never touches chunks pays no fold cost. An element
// splice changes the fingerprint and rebuilds both plans.
func (renderer *Renderer) tilePlanFor(d *layout.Display) *chunkPlan {
	if d == nil {
		return nil
	}

	gen := displayFingerprint(d)
	root := renderer.planFor(d)

	renderer.plans.mu.Lock()
	defer renderer.plans.mu.Unlock()

	for i := range renderer.plans.entries {
		e := &renderer.plans.entries[i]
		if e.d == d && e.gen == gen {
			if e.plan == nil || e.plan.gen != gen {
				e.plan = buildChunks(d, root)
			}

			return e.plan
		}
	}

	return buildChunks(d, root)
}

// tileChunkRaster returns the baked image for one run, reusing the signature
// computed at plan build while the op-write generation still matches. A tick
// that rewrites geometry, colour, opacity, or transform changes the write
// generation, recomputes the signature once, and refreshes the cache; a
// repaint with no writes reuses without rehashing.
func (renderer *Renderer) tileChunkRaster(d *layout.Display, p *chunkPlan, c *chunkRun) *chunkEntry {
	sig, box, hit := cachedChunkSig(d, c, p.content)
	if box.Empty() {
		return nil
	}

	key := chunkKey{id: c.id}

	e := renderer.chunks.get(key, sig, box, func(img *ebiten.Image) {
		for _, op := range c.ops {
			countOp()
			renderer.drawOp(img, op, -float64(box.Min.X), -float64(box.Min.Y))
		}
	})

	if !hit {
		refreshRunSig(c, sig, box)
	}

	return e
}

// paintTileRun paints one chunk when the pick accepts any member, using the
// cached raster. A mixed translation, a fractional one, or a translucent
// member keeps the per-op path, matching paintRun.
func (renderer *Renderer) paintTileRun(dst *ebiten.Image, d *layout.Display, p *chunkPlan, c *chunkRun, pick opPicker) {
	tx, ty, picked, shared := pickRun(c, pick)
	if !picked {
		return
	}

	if !shared || tx != math.Trunc(tx) || ty != math.Trunc(ty) || !runOpaque(c) {
		renderer.drawRunDirect(dst, c, pick)

		return
	}

	e := renderer.tileChunkRaster(d, p, c)
	if e == nil {
		renderer.drawRunDirect(dst, c, pick)

		return
	}

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(e.box.Min.X)+tx, float64(e.box.Min.Y)+ty)
	countBlit()
	dst.DrawImage(e.img, opts)
}

// paintTileRects paints each dirty rect once, gathering candidates from the
// plan's tile buckets instead of scanning the whole plan per tile. rects are
// in CSS pixels in the page's space; dx and dy move the paint as in Draw.
// Fixed ops paint only when includeFixed is set. Each rect paints into its
// own sub-image so a large op is clipped to the rect, matching DrawRect.
func (renderer *Renderer) paintTileRects(dst *ebiten.Image, d *layout.Display, p *chunkPlan, rects []image.Rectangle, dx, dy float64, includeFixed bool) {
	if dst == nil || d == nil || p == nil || len(rects) == 0 {
		return
	}

	ppt := d.PointsPerPixel
	seen := make([]int, len(p.items))
	gen := 0

	for _, rect := range rects {
		clipped := rect.Intersect(p.content).Intersect(dst.Bounds())
		if clipped.Empty() {
			continue
		}

		clip, ok := dst.SubImage(clipped).(*ebiten.Image)
		if !ok {
			continue
		}

		gen++
		var cands []int
		if p.tiles != nil {
			cands = p.tiles.planTilesForRect(clipped, seen, gen, cands[:0])
			sortPlanCandidates(cands)
		} else {
			cands = cands[:0]
			for i := range p.items {
				cands = append(cands, i)
			}
		}

		pick := func(op *layout.DisplayOp) (float64, float64, bool) {
			if !includeFixed && op.Fixed {
				return 0, 0, false
			}

			if !opTouches(op, ppt, clipped) {
				return 0, 0, false
			}

			return dx, dy, true
		}

		for _, at := range cands {
			it := &p.items[at]

			switch {
			case it.group != nil:
				renderer.paintGroup(clip, d, it.group, pick)
			case it.run != nil:
				renderer.paintTileRun(clip, d, p, it.run, pick)
			case it.op != nil:
				if ox, oy, ok := pick(it.op); ok {
					countOp()
					renderer.drawOp(clip, it.op, ox, oy)
				}
			}
		}
	}
}
