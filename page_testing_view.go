package ownframe

import (
	"context"
	"image"
)

// PNG encodes the page to PNG bytes.
func (t Testing) PNG() []byte {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.PNG()
}

// PNGWithContext threads ctx to the on-demand render.
func (t Testing) PNGWithContext(ctx context.Context) []byte {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.PNGWithContext(ctx)
}

// Display returns the last display list, or nil for a rasterized page.
func (t Testing) Display() *Display {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Display()
}

// Image returns the last rasterized picture, or nil for a display-list page.
func (t Testing) Image() image.Image {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Image()
}
