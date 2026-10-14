package replay

import (
	"context"
	"fmt"
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/render"
)

// textParityOps are the runs the parity check draws both ways: a plain run, a
// letter-spaced run, a fake-bold run, a half-alpha color, the plain run again
// twelve whole pixels lower and 120 right, which must share its raster, and
// one run too wide to cache, which must draw directly and not enter the cache.
// The shifted run moves by whole pixels on both axes, so its key must match
// the plain run's; the rows sit far enough apart that no two runs overlap.
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
