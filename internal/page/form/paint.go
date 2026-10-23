package form

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/textrun"
)

// PaintCaret appends the focused field's caret when the range is empty.
// A missing focus, a select-all, a non-editable control, a file input, or
// a preedit that still covers a range paints nothing. blinkOn chooses a
// fill or a no-op so a blink can hide the caret without moving the text.
func PaintCaret(d *layout.Display, v FocusView, editable, file, blinkOn bool) {
	if v.FocusID == "" || v.All || !editable || file {
		return
	}

	n := RuneLen(v.Control.Value)
	start, end := RangeOf(v.Anchor, v.Caret, n)
	if start != end {
		return
	}

	text, offset, ok := CaretText(v)
	if !ok {
		return
	}

	box, ok := CaretBox(d, v.Control.ID)
	if !ok {
		return
	}

	sample, x, ok := textrun.Caret(d, box, text, offset)
	if !ok {
		sample, x = EmptyCaret(d, box)
	}

	if sample.Size <= 0 {
		sample.Size = CaretFontSize
	}

	sample.Kind = layout.DisplayOpNoop
	if blinkOn {
		sample.Kind = layout.DisplayOpFillRect
	}

	sample.X, sample.Y = x, sample.Y-sample.Size
	sample.W, sample.H = d.PointsPerPixel, sample.Size
	sample.Text, sample.Font = "", nil
	d.Ops = append(d.Ops, sample)
	d.Order = append(d.Order, len(d.Ops)-1)
}

// OffsetAt turns a point into a rune offset in c's value.
// A run that cannot be measured, or one whose drawn text does not line up
// with the shown value, lands at the end of the value.
func OffsetAt(d *layout.Display, box layout.Box, c Control, x, y float64) int {
	n := RuneLen(c.Value)
	off, ok := textrun.At(d, box, x, y, ShownText(strings.ToLower(c.Type), c.Value))
	if !ok {
		return n
	}

	return ClampPos(off, n)
}
