package window

import (
	"fmt"
	"image/color"

	"github.com/chinmay-sawant/blinkless/layout"
)

// devOpTextCap caps the quoted text runes in an Ops row.
const devOpTextCap = 40

// devOpRow builds one Ops-tab row for the operation at index i, the nth
// painted operation.
func devOpRow(op *layout.DisplayOp, i, n int, ink color.RGBA, ppt float64) devRow {
	x := devOpRect(op, ppt)
	sp := []devSpan{
		{text: fmt.Sprintf("#%d ", n), ink: devDim},
		{text: devOpsListName[op.Kind], ink: ink},
		{text: fmt.Sprintf(" x%g y%g %gx%g", x.X, x.Y, x.W, x.H), ink: devDim},
	}
	if op.Kind == layout.DisplayOpText || op.Kind == layout.DisplayOpBullet {
		t := op.Text
		if len([]rune(t)) > devOpTextCap {
			t = string([]rune(t)[:devOpTextCap]) + "..."
		}
		sp = append(sp, devSpan{text: fmt.Sprintf(" %q", t), ink: devStrInk})
	}

	return devRow{line: devLine{spans: sp}, act: devActOp, arg: i}
}
