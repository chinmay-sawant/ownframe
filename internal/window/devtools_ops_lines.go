package window

import (
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
)

// devOpsListName maps kind index to label; 0 and 6 are empty.
var devOpsListName = [...]string{"", "fill", "stroke", "line", "text", "image", "", "text", "grid"}

// devOpsRows returns the Ops tab rows.
func (s *shell) devOpsRows() []devRow {
	mark := "[ ]"
	if s.dev.ops {
		mark = "[x]"
	}
	r := []devRow{{devLine{[]devSpan{{mark, devAccent, 0}, {" Show paint outlines (o)", devFg, 0}}}, devActOpsToggle, 0, ""}}
	h := func(t string) { r = append(r, devRow{}, devRow{line: devText(t, devDim)}) }
	if s.display == nil {
		h("bitmap fallback")
		return r
	}

	h("DISPLAY")
	c := devCountOps(s.display)
	v := [6]int{c.Fill, c.Stroke, c.Line, c.Text, c.Image, c.Grid}
	k := [...]layout.DisplayKind{1, 2, 3, 4, 5, 8}
	for i, kind := range k {
		ink, _ := devOpInk(kind)
		lbl, num := devFg, devNumInk
		if v[i] == 0 {
			lbl, num = devDim, devDim
		}
		r = append(r, devRow{devLine{[]devSpan{devSwatch(ink), {" " + devOpsListName[kind], lbl, 0}, {fmt.Sprint(" ", v[i]), num, 0}}}, 0, 0, ""})
	}
	r = append(r, devRow{devLine{[]devSpan{{"total", devDim, 0}, {fmt.Sprint(" ", c.total()), devNumInk, 0}}}, 0, 0, ""})

	h("PAINT ORDER")
	n := 0
	for _, i := range s.display.Order {
		if i < 0 || i >= len(s.display.Ops) {
			continue
		}
		op := &s.display.Ops[i]
		ink, ok := devOpInk(op.Kind)
		if !ok {
			continue
		}
		n++
		x := devOpRect(op, s.display.PointsPerPixel)
		sp := []devSpan{
			{fmt.Sprintf("#%d ", n), devDim, 0},
			{devOpsListName[op.Kind], ink, 0},
			{fmt.Sprintf(" x%g y%g %gx%g", x.X, x.Y, x.W, x.H), devDim, 0},
		}
		if op.Kind == layout.DisplayOpText || op.Kind == layout.DisplayOpBullet {
			t := op.Text
			if len([]rune(t)) > 40 {
				t = string([]rune(t)[:40]) + "..."
			}
			sp = append(sp, devSpan{fmt.Sprintf(" %q", t), devStrInk, 0})
		}
		r = append(r, devRow{devLine{sp}, devActOp, i, ""})
	}
	if n == 0 {
		r = append(r, devRow{line: devText("(none)", devDim)})
	}
	return r
}
