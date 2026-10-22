package ownframe

import (
	"context"
	"image"
)

// PNG encodes the page to PNG bytes.
func (p *Page) PNG() []byte { return innerPage(p).PNG() }

// PNGWithContext threads ctx to the on-demand render.
func (p *Page) PNGWithContext(ctx context.Context) []byte {
	return innerPage(p).PNGWithContext(ctx)
}

// Boxes returns a copy of the last hit-test boxes.
func (p *Page) Boxes() []Box { return innerPage(p).Boxes() }

// Display returns the last display list, or nil for a rasterized page.
func (p *Page) Display() *Display { return innerPage(p).Display() }

// Image returns the last rasterized picture, or nil for a display-list page.
func (p *Page) Image() image.Image { return innerPage(p).Image() }
