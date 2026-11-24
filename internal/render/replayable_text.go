package render

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// replayableText reports whether a text run can replay on the display list.
// A quarter-turn glyph rotation (vertical writing) and unprojected horizontal
// synthetic oblique replay through the shared text matrix, OpenType features
// (font-variant, font-feature-settings) apply to the Ebiten face, and
// text-autospace gaps are added between pieces at ideograph and Latin
// boundaries. Any other rotation or a rotated/projected synthetic oblique keep
// the software painter.
func replayableText(op *layout.DisplayOp) bool {
	if op.Font == nil || !masking.ReplayableTextRotation(op) {
		return false
	}
	if op.FakeOblique {
		_, ok := masking.TextMatrix(op)

		return ok
	}

	return true
}
