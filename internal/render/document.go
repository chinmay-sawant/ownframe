package render

import (
	"context"
	"fmt"
	"image"

	"github.com/chinmay-sawant/blinkless/css"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

// DisplayListDocument turns an already styled document into vector
// operations. It is the layout half of DisplayListState, without the parse
// and the cascade.
func DisplayListDocument(
	ctx context.Context,
	styled *css.Document,
	images func(src string) ([]byte, error),
) (*Display, error) {
	return DisplayListDocumentOptions(ctx, styled, layout.Options{Images: images})
}

// DisplayListDocumentOptions turns an already styled document into vector
// operations with the given engine options. The RetainState option keeps the
// layout state a single element patch needs.
func DisplayListDocumentOptions(
	ctx context.Context,
	styled *css.Document,
	options layout.Options,
) (*Display, error) {
	if ctx == nil {
		return nil, host.ErrNilContext
	}
	engineMu.Lock()
	defer engineMu.Unlock()
	display, err := layout.DisplayListOptions(ctx, styled, options)
	if err != nil {
		return nil, fmt.Errorf("render: display: %w", err)
	}

	return display, nil
}

// ReleaseDisplay returns a superseded display list to the engine pool under
// engineMu, synchronizing the caller's prior reads of display.Boxes and
// display.Ops with the next DisplayListOptions call that reuses the pooled
// backing arrays.
func ReleaseDisplay(d *Display) {
	if d == nil {
		return
	}
	engineMu.Lock()
	d.Release()
	engineMu.Unlock()
}

// PaintDocument lays out an already styled document. It is the layout half
// of PaintState, without the parse and the cascade. The image is canvas-sized
// and blank. bitmap.Paint fills the pixels for Page.PNG and fallback pages,
// and internal/replay fills them for the window.
func PaintDocument(
	ctx context.Context,
	styled *css.Document,
	images func(src string) ([]byte, error),
) (image.Image, []layout.Box, error) {
	if ctx == nil {
		return nil, nil, host.ErrNilContext
	}
	display, err := DisplayListDocument(ctx, styled, images)
	if err != nil {
		return nil, nil, err
	}

	return canvasImage(display), display.Boxes, nil
}
