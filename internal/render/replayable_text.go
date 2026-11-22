package render

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// replayableText reports whether a text run can replay on the display list.
// A quarter-turn glyph rotation (vertical writing) replays through the shared
// text matrix, OpenType features (font-variant, font-feature-settings) apply
// to the Ebiten face, and text-autospace gaps are added between pieces at
// ideograph and Latin boundaries. Any other rotation and a synthetic oblique
// keep the software painter.
func replayableText(op *layout.DisplayOp) bool {
	return op.Font != nil && !op.FakeOblique && masking.ReplayableTextRotation(op)
}
