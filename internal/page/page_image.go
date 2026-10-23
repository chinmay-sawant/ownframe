package page

import (
	"context"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// Image returns the last rasterized picture, or nil when nothing has been
// drawn or the page is drawn from its display list.
func (p *Page) Image() image.Image {
	return p.draw.Img
}

// Display returns the last display list, or nil when the page is drawn as a
// rasterized picture.
func (p *Page) Display() *layout.Display {
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
