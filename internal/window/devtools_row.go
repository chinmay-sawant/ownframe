package window

import (
	"image/color"
)

// devSpan is one run on a panel line: text in one ink, or a filled swatch
// when swatch is positive. w is the measured text width when the builder
// knows it; a zero w measures while drawing. A right span starts at the
// line's right edge, which needs its w.
type devSpan struct {
	text   string
	ink    color.RGBA
	swatch float64
	w      float64
	right  bool
}

// devLine is one panel line, drawn left to right.
type devLine struct {
	spans []devSpan
}

// devAct is what a panel row does when a click lands on it.
type devAct int

const (
	devActNone      devAct = iota
	devActTab              // select the tab named by arg
	devActOpsToggle        // flip the operation outlines
	devActJSON             // collapse or expand the JSON node named by key
	devActOp               // outline the operation at index arg
)

// devRow is one content line plus the click action it carries. w is the line
// width the builder measured against, for right-aligned spans.
type devRow struct {
	line devLine
	act  devAct
	arg  int
	key  string
	w    float64
}

// devSwatchSize is the side of a colour chip in a panel line.
const devSwatchSize = 9.0

// devSwatchGap is the space between a swatch and the text after it.
const devSwatchGap = 6.0

// devText builds a one-span line.
func devText(s string, ink color.RGBA) devLine {
	return devLine{spans: []devSpan{{text: s, ink: ink}}}
}

// devPlain builds a line in the default panel ink.
func devPlain(s string) devLine {
	return devText(s, devFg)
}

// devSwatch builds a colour chip span.
func devSwatch(ink color.RGBA) devSpan {
	return devSpan{ink: ink, swatch: devSwatchSize}
}

// plain is the line without colour, for labels and tests.
func (l devLine) plain() string {
	out := ""
	for _, span := range l.spans {
		out += span.text
	}

	return out
}
