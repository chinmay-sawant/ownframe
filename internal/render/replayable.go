package render

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// Replayable reports whether the replay painter can draw every operation in
// display like the bitmap painter: same order, colors, glyphs, and
// compositing. A false result keeps the page on the bitmap path.
//
// A CSS outline is an ordinary line or stroke op that Display.Order moves to
// the outline paint layer, so outlines replay. A finite 2D transform replays
// on every kind the painter draws from geometry, and a blend or isolation
// group replays when its mode is separable and it stays inside one paint band.
func Replayable(display *layout.Display) bool {
	if display == nil {
		return false
	}
	if masking.ValidateDisplay(display) != nil {
		return false
	}

	for i := range display.Ops {
		if !replayableOp(&display.Ops[i], display.PixelPerPoint) {
			return false
		}
	}

	return true
}

func replayableOp(op *layout.DisplayOp, ppp float64) bool {
	if op.GroupBoundary() != 0 {
		return replayableGroup(op)
	}

	if !replayableGroup(op) {
		return false
	}

	if op.XformSet {
		if !finiteXform(op) || op.Kind == layout.DisplayOpGridRun {
			return false
		}

		if op.Kind != layout.DisplayOpImage && xformTooBig(op, ppp) {
			return false
		}
	}

	switch op.Kind {
	case layout.DisplayOpNoop, layout.DisplayOpLinkURI:
		return true
	case layout.DisplayOpFillRect:
		_, ok := FillRadii(op)

		return ok
	case layout.DisplayOpStrokeRect:
		// Top, right, bottom, and left are the only defined mask bits.
		return op.StrokeMask&^0x0F == 0
	case layout.DisplayOpImage:
		data, _, _ := op.ImageBytes()

		return data != nil
	case layout.DisplayOpLine:
		return true
	case layout.DisplayOpText, layout.DisplayOpBullet:
		return replayableText(op)
	case layout.DisplayOpGridRun:
		return op.Grid != nil
	default:
		return false
	}
}
