package form

import "github.com/chinmay-sawant/blinkless/layout"

// caretXOffset nudges the empty caret right of the box edge.
const caretXOffset = 7

// caretHeightRatio drops the empty caret top to this share of the box.
const caretHeightRatio = 0.65

// CaretFontSize is the fallback caret size when the sample has none.
const CaretFontSize = 12

// CaretBox is the display box whose id matches.
func CaretBox(d *layout.Display, id string) (layout.Box, bool) {
	for _, box := range d.Boxes {
		if box.ID == id {
			return box, true
		}
	}

	return layout.Box{}, false
}

// EmptyCaret is the caret sample and x position when the field has no text run.
func EmptyCaret(d *layout.Display, box layout.Box) (layout.DisplayOp, float64) {
	pt := d.PointsPerPixel

	return layout.DisplayOp{
		Y:     (box.Y + box.H*caretHeightRatio) * pt,
		Size:  CaretFontSize,
		Alpha: 1,
	}, (box.X + caretXOffset) * pt
}
