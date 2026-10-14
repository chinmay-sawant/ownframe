package replay

import (
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
)

// paintChunked replays display in paint order with element runs served from
// the chunk cache. It falls back to the direct walk when no plan is cached.
func paintChunked(dst *ebiten.Image, d *layout.Display, pick opPicker) {
	if dst == nil || d == nil {
		return
	}

	p := chunkPlanFor(d)
	if p == nil {
		paintDisplay(dst, d, pick)

		return
	}

	for i := range p.items {
		it := &p.items[i]

		switch {
		case it.group != nil:
			paintGroup(dst, d, it.group, pick)
		case it.run != nil:
			paintRun(dst, d, p, it.run, pick)
		case it.op != nil:
			if dx, dy, ok := pick(it.op); ok {
				countOp()
				drawOp(dst, it.op, dx, dy)
			}
		}
	}
}

// paintRun paints one chunk when the pick accepts any member, at the
// translation the picked members share. A mixed translation, a fractional
// one, or a translucent member keeps the per-op path: an image blit can only
// copy whole pixels, and a translucent op cannot promise byte parity.
func paintRun(dst *ebiten.Image, d *layout.Display, p *chunkPlan, c *chunkRun, pick opPicker) {
	tx, ty, picked, shared := pickRun(c, pick)
	if !picked {
		return
	}

	if !shared || tx != math.Trunc(tx) || ty != math.Trunc(ty) || !runOpaque(c) {
		drawRunDirect(dst, c, pick)

		return
	}

	e := chunkRaster(d, p, c)
	if e == nil {
		drawRunDirect(dst, c, pick)

		return
	}

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(e.box.Min.X)+tx, float64(e.box.Min.Y)+ty)
	countBlit()
	dst.DrawImage(e.img, opts)
}
