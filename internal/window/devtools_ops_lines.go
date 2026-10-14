package window

import (
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
)

// devOpsListName is the per-kind label.
var devOpsListName = [...]string{"", "fill", "stroke", "line", "text", "image", "", "text", "grid"}

// devOpsBuildRows builds the cached Ops rows.
func (s *shell) devOpsBuildRows() []devRow {
	mark := "[ ]"
	if s.dev.ops {
		mark = "[x]"
	}
	r := []devRow{{line: devLine{spans: []devSpan{{text: mark, ink: devAccent}, {text: " Show paint outlines (o)", ink: devFg}}}, act: devActOpsToggle}}
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
		r = append(r, devRow{line: devLine{spans: []devSpan{devSwatch(ink), {text: " " + devOpsListName[kind], ink: lbl}, {text: fmt.Sprint(" ", v[i]), ink: num}}}})
	}
	r = append(r, devRow{line: devLine{spans: []devSpan{{text: "total", ink: devDim}, {text: fmt.Sprint(" ", c.total()), ink: devNumInk}}}})

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
		r = append(r, devOpRow(op, i, n, ink, s.display.PointsPerPixel))
	}
	if n == 0 {
		r = append(r, devRow{line: devText("(none)", devDim)})
	}
	return r
}
