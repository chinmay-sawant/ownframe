// Package render reads a placement from layout.DisplayList. blinkless returns
// that list and no page bitmap. Paint returns the hit boxes and a canvas-sized
// image. Fallback pages and Page.PNG fill pixels with bitmap.Paint, and the
// window fills them with internal/replay. DisplayList keeps the operations.
// This package never imports replay; replay imports render. It does not
// build a PDF.
package render

import (
	"context"
	"image"

	"github.com/chinmay-sawant/blinkless/layout"
)

// Paint parses source and returns the screen picture and hit boxes.
func Paint(ctx context.Context, source string, width, height int) (image.Image, []layout.Box, error) {
	return PaintState(ctx, source, width, height, State{})
}

// PaintState is Paint with the runtime pointer and focus state.
func PaintState(ctx context.Context, source string, width, height int, state State) (image.Image, []layout.Box, error) {
	display, err := DisplayListState(ctx, source, width, height, state)
	if err != nil {
		return nil, nil, err
	}

	return canvasImage(display), display.Boxes, nil
}

// canvasImage is a blank image the size of the display. bitmap.Paint fills
// the pixels for Page.PNG and fallback pages, and internal/replay fills
// them for the window; neither direction imports back into render.
func canvasImage(display *Display) image.Image {
	if display == nil || display.Width < 1 || display.Height < 1 {
		return nil
	}

	return image.NewNRGBA(image.Rect(0, 0, display.Width, display.Height))
}
