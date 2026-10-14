package replay

import (
	"fmt"
	"math"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// drawTextDirect draws one text op through the uncached path only.
func drawTextDirect(op *layout.DisplayOp, dst *ebiten.Image) {
	face := textFace(op)
	if face == nil || op.Text == "" {
		return
	}

	content := layout.DisplayTransformText(op.Text, op.TextTransformValue())
	drawTextRun(dst, face, content, op.X*pxPerPt, op.Y*pxPerPt-face.Metrics().HAscent, op)
}

// phase is the run's keyed sub-pixel offset, built the way cachedTextRun
// builds it from the op and the face.
func phase(op *layout.DisplayOp, face *text.GoTextFace) [2]float64 {
	x := op.X * pxPerPt
	y := op.Y*pxPerPt - face.Metrics().HAscent

	return [2]float64{textPhase(x - math.Floor(x)), textPhase(y - math.Floor(y))}
}

// setRunBounds fills the advance and ink metrics the dirty-rect filter reads,
// so the clipped parity check draws the runs the way a real page would.
func setRunBounds(ops []layout.DisplayOp, base *layout.DisplayOp) error {
	face := textFace(base)
	if face == nil {
		return fmt.Errorf("the parity runs have no face")
	}

	m := face.Metrics()
	for i := range ops {
		width, height := text.Measure(ops[i].Text, face, 0)
		ops[i].W, ops[i].H = width/pxPerPt, height/pxPerPt
		ops[i].InkDescent = m.HDescent / pxPerPt
	}

	return nil
}
