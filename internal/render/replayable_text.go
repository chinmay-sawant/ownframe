package render

import "github.com/chinmay-sawant/blinkless/layout"

func replayableText(op *layout.DisplayOp) bool {
	if op.Font == nil || op.RotateDeg != 0 || op.FakeOblique {
		return false
	}
	return op.FontFeatures() == "" && op.TextAutospaceGap() == 0
}
