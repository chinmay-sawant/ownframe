package page

import (
	"bytes"
	"context"
	"image"
	"image/png"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// Image returns the last rasterized picture, or nil when nothing has been
// drawn or the page is drawn from its display list.
func (p *Page) Image() image.Image {
	return p.img
}

// Display returns the last display list, or nil when the page is drawn as a
// rasterized picture.
func (p *Page) Display() *layout.Display {
	return p.display
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
	if p.png != nil {
		return append([]byte(nil), p.png...)
	}

	img := p.img
	if img == nil && p.display != nil {
		var err error
		img, err = bitmap.Paint(p.display)
		if err != nil || img == nil {
			return nil
		}
	}

	if img == nil {
		if p.source == "" {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return nil
		}
		display, err := render.DisplayListState(ctx, p.source, p.width, p.height, p.renderState())
		if err != nil {
			return nil
		}

		img, err = bitmap.Paint(display)
		if err != nil || img == nil {
			return nil
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}

	p.png = buf.Bytes()

	return append([]byte(nil), p.png...)
}

// Boxes returns a copy of the last hit-test boxes, or nil when nothing has
// been drawn. Caller edits never corrupt the page index.
func (p *Page) Boxes() []Box {
	return append([]Box(nil), p.boxes...)
}
