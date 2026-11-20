package page

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/motion"
	"github.com/chinmay-sawant/ownframe/internal/page/draw"
	"image"
)

// Image returns the last rasterized picture, or nil when nothing has been
// drawn or the page is drawn from its display list.
func (p *Page) Image() image.Image {
	return p.draw.Img
}

// Display returns the last display list, or nil when the page is drawn as a
// rasterized picture.
func (p *Page) Display() *layout.Display {
	if p.tickActive && p.tickSnap == nil {
		p.tickSnap = p.captureTickOps()
		p.tickMutated.Store(true)
	}
	return p.draw.Display
}

// PNG encodes the page to PNG bytes. A display-list page rasterizes on
// demand, once, and the bytes are cached until the next paint change. It returns
// a copy, so caller edits never corrupt the cache, and nil when the page
// cannot be painted.
func (p *Page) PNG() []byte {
	return p.PNGWithContext(context.Background())
}

// PNGWithContext threads ctx to the on-demand render.
func (p *Page) PNGWithContext(ctx context.Context) []byte {
	if ctx == nil {
		ctx = context.Background()
	}

	return p.draw.Bytes(ctx, p.source, p.viewport(), p.renderState())
}

// SetImage registers encoded image bytes (PNG, JPEG, or SVG) for one source
// name. A template image source that spells that name, such as
// background-image: url("logo") or <img src="logo">, resolves to data on the
// next Redraw. The bytes are cloned, so later caller edits do not repaint
// the page. Passing nil data removes the entry.
func (p *Page) SetImage(src string, data []byte) {
	p.images = draw.Store(p.images, src, data)
	p.imagesRev++
}

// ImageData returns a copy of the bytes registered for src.
// A src the page never registered falls back to the bundled emoji pictures.
// The web preview serves this for url("name") and img src="name".
func (p *Page) ImageData(src string) ([]byte, bool) {
	if p == nil {
		return nil, false
	}

	data, err := p.imageBytes(src)
	if err != nil || len(data) == 0 {
		return nil, false
	}

	return append([]byte(nil), data...), true
}

// imageBytes resolves one image source for the layout engine. A src the
// page never registered falls back to the bundled emoji pictures, so the
// replacer needs no per-page setup. Blocked schemes (GEN-30) fail before
// any lookup.
func (p *Page) imageBytes(src string) ([]byte, error) {
	if err := checkImageSrc(src); err != nil {
		return nil, err
	}

	data, ok := draw.Lookup(p.images, src)
	if !ok {
		return nil, errNoImage
	}

	return data, nil
}

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
	if previous != nil && previous != d {
		previous.Release()
	}
	p.markRect(patch.Old)
	if patch.New != patch.Old {
		p.markRect(patch.New)
	}

	return true
}
