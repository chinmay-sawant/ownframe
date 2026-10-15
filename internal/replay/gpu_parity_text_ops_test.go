package replay

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// textParityOps are the runs drawn both ways: plain, spaced, bold, dim, a
// whole-pixel shifted share, and one too wide to cache.
func textParityOps() ([]layout.DisplayOp, error) {
	d, err := render.DisplayList(context.Background(),
		`<html><body style="margin:0;font-size:14px"><p>parity</p></body></html>`, 320, 240)
	if err != nil {
		return nil, err
	}

	found := firstTextOp(d)
	if found == nil {
		return nil, fmt.Errorf("no text op in the fixture")
	}

	plain := *found
	plain.Text, plain.Size = "Raster run 128", 14
	plain.Bold, plain.LetterSpacing, plain.Alpha = false, 0, 0
	plain.X, plain.Y = 3, 60
	plain.R, plain.G, plain.B = 0.1, 0.2, 0.3

	spaced := plain
	spaced.Text, spaced.LetterSpacing, spaced.Y = "spaced run", 3.5, 78

	fakeBold := plain
	fakeBold.Text, fakeBold.Bold, fakeBold.Y = "Fake bold 17", true, 96

	dim := plain
	dim.Text, dim.Alpha, dim.Y = "Half alpha 9", 0.5, 114

	shifted := plain
	shifted.X += 90
	shifted.Y += 72

	wide := plain
	wide.Text, wide.Y = strings.Repeat("wide ", 40), 150

	ops := []layout.DisplayOp{plain, spaced, fakeBold, dim, shifted, wide}

	return ops, setRunBounds(ops, &plain)
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
