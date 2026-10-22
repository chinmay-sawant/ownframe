package draw

import (
	"context"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/bitmap"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// Picture paints d to a bitmap. A nil or empty list returns nil, nil.
func Picture(d *layout.Display) (image.Image, error) {
	return bitmap.Paint(d)
}

// PaintSource lays source out and paints it. PNG uses it when the page
// has no retained picture yet.
func PaintSource(
	ctx context.Context,
	source string,
	w, h int,
	state render.State,
) (image.Image, error) {
	display, err := render.DisplayListState(ctx, source, w, h, state)
	if err != nil {
		return nil, err
	}

	return bitmap.Paint(display)
}
