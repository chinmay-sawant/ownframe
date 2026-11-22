package render

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/masking"
)

// replayableText reports whether a text run can replay on the display list.
// A quarter-turn glyph rotation (vertical writing) replays through the shared
// text matrix; any other rotation, a synthetic oblique, font features, and
// autospace gaps keep the software painter.
func replayableText(op *layout.DisplayOp) bool {
	if op.Font == nil || op.FakeOblique || !masking.ReplayableTextRotation(op) {
		return false
	}
	return op.FontFeatures() == "" && op.TextAutospaceGap() == 0
}
