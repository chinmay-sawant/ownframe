// Package replay draws a layout display list onto an Ebiten image. The window
// uses it to show the page as vector operations instead of one rasterized
// bitmap. The caller decides replayability first; a page it cannot draw keeps
// the bitmap path.
package replay

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"image/color"
	"math"
)

// Draw replays display in paint order. dx and dy move the page's top-left in
// pixels, so the window passes its scroll offset. A viewport-fixed op ignores
// the offset: the engine placed it against the viewport. Blend and isolation
// groups composite once through the paint plan.
func (renderer *Renderer) Draw(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	renderer.paintDisplay(dst, display, func(*layout.DisplayOp) (float64, float64, bool) {
		return dx, dy, true
	})
}

// DrawUnfixed replays only the operations that scroll with the page. A
// caller that paints a scrolled buffer draws the fixed ops separately, after
// the buffer, so they stay at the viewport.
func (renderer *Renderer) DrawUnfixed(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	renderer.paintDisplay(dst, display, func(op *layout.DisplayOp) (float64, float64, bool) {
		return dx, dy, !op.Fixed
	})
}

// DrawFixed replays only the viewport-fixed operations, untranslated.
func (renderer *Renderer) DrawFixed(dst *ebiten.Image, display *layout.Display) {
	renderer.paintDisplay(dst, display, func(op *layout.DisplayOp) (float64, float64, bool) {
		return 0, 0, op.Fixed
	})
}

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
			} else if e.plan.geomGen != displayGeomGen(d) {
				refreshPlanTiles(d, e.plan)
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

// viewportUnionFactor bounds the overdraw a union rect may add before the
// picker tests each rect instead. A union at most half again the summed area
// paints little outside the rects, so one bounds check per op beats one per
// rect; a sprawling union falls back to per-rect checks.
const viewportUnionFactor = 3

// viewportUnion returns the bounding union of rects and whether one bounds
// check against it covers the set. Empty rects are skipped. A single rect
// always uses the union fast path. For several rects, the union wins when its
// area is below the threshold: union*2 <= sum*3, so at most half again the
// precise area paints.
func viewportUnion(rects []image.Rectangle) (image.Rectangle, bool) {
	kept := rects[:0]

	for _, r := range rects {
		if !r.Empty() {
			kept = append(kept, r)
		}
	}

	rects = kept

	if len(rects) == 0 {
		return image.Rectangle{}, false
	}

	if len(rects) == 1 {
		return rects[0], true
	}

	union := rects[0]
	var sum int64

	for _, r := range rects {
		union = union.Union(r)
		sum += int64(r.Dx()) * int64(r.Dy())
	}

	unionArea := int64(union.Dx()) * int64(union.Dy())
	if unionArea*2 <= sum*viewportUnionFactor {
		return union, true
	}

	return union, false
}

// opTouchesAny reports whether an op's painted box meets any rect, with one
// bounds lookup when useUnion holds and one per rect otherwise. An op whose
// ink cannot be bounded always touches, matching opTouches.
func opTouchesAny(op *layout.DisplayOp, ppt float64, rects []image.Rectangle, union image.Rectangle, useUnion bool) bool {
	if useUnion {
		return opTouches(op, ppt, union)
	}

	for _, r := range rects {
		if opTouches(op, ppt, r) {
			return true
		}
	}

	return false
}

// drawOp replays one operation. Kinds that carry no paint are ignored.
// A viewport-fixed op ignores the scroll translation: the engine placed it
// against the viewport, so it stays on screen while the page moves. An op
// with a non-identity transform draws through its baked matrix; an image
// already places itself through its own transform, and a corner path maps
// each command point through the matrix as it builds the vector path.
func (renderer *Renderer) drawOp(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	if op.Fixed {
		dx, dy = 0, 0
	}
	renderer.drawOpAt(dst, op, dx, dy)
}

// drawOpInLayer keeps the layer-local translation for fixed ops. The
// enclosing group already removed the page scroll and restores its origin
// when it composites this layer.
func (renderer *Renderer) drawOpInLayer(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	renderer.drawOpAt(dst, op, dx, dy)
}

func (renderer *Renderer) drawOpAt(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	if _, ok := textProjection(op); ok {
		renderer.drawXformText(dst, op, dx, dy)
		return
	}

	if op.XformSet && op.Kind != layout.DisplayOpImage && op.Kind != layout.DisplayOpPath &&
		!op.Transform().IsIdentity() {
		renderer.drawXform(dst, op, dx, dy)

		return
	}

	switch op.Kind {
	case layout.DisplayOpFillRect:
		renderer.fillRect(dst, op, dx, dy)
	case layout.DisplayOpStrokeRect:
		renderer.drawStrokeRect(dst, op, dx, dy)
	case layout.DisplayOpPath:
		drawPath(dst, op, dx, dy)
	case layout.DisplayOpImage:
		renderer.drawImage(dst, op, dx, dy)
	case layout.DisplayOpLine:
		drawLine(dst, op, dx, dy)
	case layout.DisplayOpGridRun:
		drawGrid(dst, op, dx, dy)
	case layout.DisplayOpText, layout.DisplayOpBullet:
		renderer.drawText(dst, op, dx, dy)
	}
}

// scissorOverflow returns dst clipped to the op's recorded overflow box.
// Geometry-baked clipping shrinks the op bounds, which is exact for static
// paint but moves with a sampled matrix; a host painting through a transform
// must intersect the paint with the ancestor clip instead. It reports false
// when the clip erases the paint so the caller skips the draw. Ops without a
// recorded clip keep dst unchanged.
func scissorOverflow(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) (*ebiten.Image, bool) {
	cx, cy, cw, ch, ok := op.OverflowClip()
	if !ok || dst == nil {
		return dst, true
	}

	r := image.Rect(
		int(math.Floor(cx*pxPerPt+dx)),
		int(math.Floor(cy*pxPerPt+dy)),
		int(math.Ceil((cx+cw)*pxPerPt+dx)),
		int(math.Ceil((cy+ch)*pxPerPt+dy)),
	).Intersect(dst.Bounds())

	if r.Empty() {
		return dst, false
	}

	if r.Eq(dst.Bounds()) {
		return dst, true
	}

	sub, ok := dst.SubImage(r).(*ebiten.Image)
	if !ok {
		return dst, true
	}

	return sub, true
}

// DrawVisible skips offscreen operations before text shaping or GPU
// submission. Unbounded ink is retained. Translation maps the target back
// into page space. Blend and isolation groups composite as a unit.
func (renderer *Renderer) DrawVisible(dst *ebiten.Image, display *layout.Display, dx, dy float64) {
	renderer.DrawVisiblePinned(dst, display, dx, dy, 0, 0, 0)
}

// DrawVisiblePinned is DrawVisible with a viewport-pinned layer: operations
// whose z-index is at least pinZ are drawn at pinDx, pinDy instead of the
// scroll translation, so they stay at the viewport without a redraw.
func (renderer *Renderer) DrawVisiblePinned(dst *ebiten.Image, display *layout.Display, dx, dy float64, pinZ int, pinDx, pinDy float64) {
	if dst == nil || display == nil {
		return
	}

	rect := visibleRect(dst.Bounds(), dx, dy)
	pinRect := visibleRect(dst.Bounds(), pinDx, pinDy)
	ppt := display.PointsPerPixel

	renderer.paintChunked(dst, display, func(op *layout.DisplayOp) (float64, float64, bool) {
		odx, ody, r := dx, dy, rect
		if pinZ > 0 && op.ZIndexSet && op.ZIndex >= pinZ {
			odx, ody, r = pinDx, pinDy, pinRect
		}

		if op.Fixed {
			// A fixed op ignores the page scroll, including when a chunk
			// run batches several fixed ops into one blit. The per-op
			// drawOp path resets the translation itself, but a run blits
			// at the translation the picker returned, so return zero here
			// as well or the chrome jumps by the scroll offset.
			r = dst.Bounds()
			odx, ody = 0, 0
		}

		if !opTouches(op, ppt, r) {
			return 0, 0, false
		}

		return odx, ody, true
	})
}

func visibleRect(bounds image.Rectangle, dx, dy float64) image.Rectangle {
	return image.Rect(
		int(math.Floor(float64(bounds.Min.X)-dx)),
		int(math.Floor(float64(bounds.Min.Y)-dy)),
		int(math.Ceil(float64(bounds.Max.X)-dx)),
		int(math.Ceil(float64(bounds.Max.Y)-dy)),
	)
}

func visitVisible(display *layout.Display, rect image.Rectangle, draw func(int)) {
	for _, index := range display.Order {
		if index < 0 || index >= len(display.Ops) {
			continue
		}
		if opTouches(&display.Ops[index], display.PointsPerPixel, rect) {
			draw(index)
		}
	}
}

// pxPerPt is the CSS pixel length of one layout point at zoom 1.
const pxPerPt = 96.0 / 72.0

// colorChannelMax is the full-scale value of a 16-bit color channel, the
// divisor that maps color.RGBA's channel range to the vector color scale.
const colorChannelMax = 0xffff

// rgba converts an op's color and effective alpha to a color.RGBA. Opacity
// folds element opacity and the op alpha exactly as the engine's painters do.
func rgba(op *layout.DisplayOp) color.RGBA {
	return masking.PaintRGBA(op)
}

// contentRect is the painted page in CSS pixels: the canvas plus any box
// that overflows it, the region the window's replay buffer covers.
func contentRect(display *layout.Display) image.Rectangle {
	r := canvasRect(display)
	if display == nil {
		return r
	}

	for _, b := range display.Boxes {
		r = r.Union(contentBox(b))
	}

	return r
}

// contentBox rounds one element box outward to whole CSS pixels.
func contentBox(b layout.Box) image.Rectangle {
	return image.Rect(
		int(math.Floor(b.X)), int(math.Floor(b.Y)),
		int(math.Ceil(b.X+b.W)), int(math.Ceil(b.Y+b.H)),
	)
}

// drawGrid replays one collapsed table grid as its line segments, in
// emission order. Each segment inherits the owner's opacity, transform, and
// fixed placement through masking.GridLine, matching the bitmap painter.
func drawGrid(dst *ebiten.Image, op *layout.DisplayOp, dx, dy float64) {
	if op.Grid == nil {
		return
	}

	for i := range op.Grid.Segs {
		line := masking.GridLine(op, i)

		drawLine(dst, &line, dx, dy)
	}
}
