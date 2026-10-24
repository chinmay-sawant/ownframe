package page

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/motion"
)

// Inline-SVG paint application.
//
// The motion runtime samples an element's fill and stroke into a
// motion.SVGPaintRequest because the paint is baked into the SVG raster, not a
// display field. The page hands at most one request per tick to the engine's
// restyle pass and adopts the patch it returns.

// stepSVGPaintMotion applies the inline-SVG restyles the last Step collected,
// under the runtime's one-per-frame budget. It runs after Step so the requests
// belong to the frame being drawn, and before the tick's dirty rects are
// consumed, because the adopted patch marks the image bounds it changed.
//
// A restyle serializes and rasterizes the SVG, so it needs the retained layout
// state the engine patches. A page that started without it turns the state on
// and owes one redraw, the same debt the re-projection pass pays; the request is
// collected again on the frame that can apply it.
func (p *Page) stepSVGPaintMotion(ctx context.Context) {
	if p.motion == nil || p.animationsOff {
		return
	}

	requests := p.motion.TakeSVGPaint()
	if len(requests) == 0 {
		return
	}
	if p.draw.Display == nil || p.cache == nil {
		return
	}
	if !p.retainState || p.restampRedraw {
		p.retainState = true
		p.restampRedraw = true

		return
	}

	applied := 0
	for i := range requests {
		if err := p.applySVGPaint(ctx, requests[i]); err != nil {
			p.motion.NoteSVGPaint(applied, err)

			return
		}
		applied++
	}

	p.motion.NoteSVGPaint(applied, nil)
}

// applySVGPaint restyles one element and adopts the engine patch. The override
// strings are parsed by the engine, so the host does not carry a second colour
// or dash parser.
func (p *Page) applySVGPaint(ctx context.Context, request motion.SVGPaintRequest) error {
	if request.ElementID == "" || len(request.Overrides) == 0 {
		return nil
	}

	patch, err := layout.RestyleInlineSVG(ctx, p.cache.Styled(), p.draw.Display,
		request.Overrides, layout.Options{RetainState: true})
	if err != nil {
		return fmt.Errorf("motion: restyle svg %q: %w", request.ElementID, err)
	}

	p.adoptSVGPaint(patch)

	return nil
}

// adoptSVGPaint installs an SVG restyle patch through the splice path and
// rebinds every player, so the player that owns the element keeps its clock and
// does not restart on its own patch. Damage comes from the patch image bounds:
// the paint changed inside the same box, so the display diff cannot tell which
// pixels moved and the engine already measured them.
func (p *Page) adoptSVGPaint(patch *layout.ElementPatch) bool {
	if patch == nil {
		return false
	}
	d := patch.Display
	if d == nil || d == p.draw.Display {
		return false
	}

	previous := p.adoptSpliced(d)
	p.motion.MarkSVGPaintKept()
	p.bindMotion(previous, d)
	p.bindTransitions(nil)
	p.markRect(patch.Old)
	if patch.New != patch.Old {
		p.markRect(patch.New)
	}

	return true
}
