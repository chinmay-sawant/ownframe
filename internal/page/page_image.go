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
// nil when the page cannot be painted.
func (p *Page) PNG() []byte {
	if p.png != nil {
		return p.png
	}

	img := p.img
	if img == nil && p.display != nil {
		img = bitmap.Picture(p.display)
		if img == nil {
			return nil
		}
	}

	if img == nil {
		if p.source == "" {
			return nil
		}

		display, err := render.DisplayListState(context.Background(), p.source, p.width, p.height, p.renderState())
		if err != nil {
			return nil
		}

		img = bitmap.Picture(display)
	}

	if img == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil
	}

	p.png = buf.Bytes()

	return p.png
}

// Boxes returns the last hit-test boxes, or nil when nothing has been drawn.
func (p *Page) Boxes() []Box {
	return p.boxes
}
