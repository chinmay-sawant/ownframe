package page

import (
	"context"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/page/draw"
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
// demand, once, and the bytes are cached until the next Redraw. It returns
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
	if b := p.draw.CopyPNG(); b != nil {
		return b
	}
	img := p.pngImage(ctx)
	if img == nil {
		return nil
	}
	p.draw.StorePNG(draw.EncodePNG(img))

	return p.draw.CopyPNG()
}

func (p *Page) pngImage(ctx context.Context) image.Image {
	img := p.draw.Img
	if img == nil && p.draw.Display != nil {
		painted, err := draw.Picture(p.draw.Display)
		if err != nil || painted == nil {
			return nil
		}

		return painted
	}
	if img != nil {
		return img
	}
	if p.source == "" || ctx.Err() != nil {
		return nil
	}
	painted, err := draw.PaintSource(ctx, p.source, p.width, p.height, p.renderState())
	if err != nil || painted == nil {
		return nil
	}

	return painted
}
