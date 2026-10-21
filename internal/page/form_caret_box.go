package page

import "github.com/chinmay-sawant/blinkless/layout"

// caretXOffset nudges the empty caret right of the box edge.
const caretXOffset = 7

// caretHeightRatio drops the empty caret top to this share of the box.
const caretHeightRatio = 0.65

// caretFontSize is the fallback caret size when the sample has none.
const caretFontSize = 12

func caretBox(d *layout.Display, id string) (layout.Box, bool) {
	for _, box := range d.Boxes {
		if box.ID == id {
			return box, true
		}
	}
	return layout.Box{}, false
}

func emptyCaret(d *layout.Display, box layout.Box) (layout.DisplayOp, float64) {
	pt := d.PointsPerPixel
	return layout.DisplayOp{Y: (box.Y + box.H*caretHeightRatio) * pt, Size: caretFontSize, Alpha: 1}, (box.X + caretXOffset) * pt
}
