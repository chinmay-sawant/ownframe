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

// MoveCaretDisplay applies one caret navigation key to c. For a textarea it
// uses the laid-out visual lines for vertical arrows (preserving desiredX) and
// line-aware Home/End, while Ctrl+Home/Ctrl+End move to the document ends.
func MoveCaretDisplay(
	d *layout.Display, box layout.Box, c Control,
	key string, caret, start, end int,
	desiredX float64, hasDesiredX bool,
) (pos int, extend bool, nextX float64, keepX bool, ok bool) {
	if c.Tag != "textarea" {
		pos, extend, ok = MoveCaret(key, c.Value, caret, start, end)

		return pos, extend, 0, false, ok
	}

	base, extend, word := Mods(key)
	n := RuneLen(c.Value)
	caret = ClampPos(caret, n)
	shown := ShownText(strings.ToLower(c.Type), c.Value)

	switch base {
	case "arrowleft", "arrowright":
		pos, extend, ok = MoveCaret(key, c.Value, caret, start, end)

		return pos, extend, 0, false, ok
	case "arrowup", "arrowdown":
		dir := -1
		if base == "arrowdown" {
			dir = 1
		}
		from := caret
		wantX := -1.0
		if start != end && !extend {
			if dir < 0 {
				from = start
			} else {
				from = end
			}
		} else if hasDesiredX {
			wantX = desiredX
		}
		if next, prefX, moved := textrun.MoveVertical(d, box, shown, from, dir, wantX); moved {
			return ClampPos(next, n), extend, prefX, true, true
		}

		return moveVerticalLogical(c.Value, from, dir), extend, 0, false, true
	case "home":
		if word {
			return 0, extend, 0, false, true
		}
		if lineStart, _, found := textrun.LineRange(d, box, shown, caret); found {
			return ClampPos(lineStart, n), extend, 0, false, true
		}
		lineStart, _ := LineBounds(c.Value, caret)

		return lineStart, extend, 0, false, true
	case "end":
		if word {
			return n, extend, 0, false, true
		}
		if _, lineEnd, found := textrun.LineRange(d, box, shown, caret); found {
			return ClampPos(lineEnd, n), extend, 0, false, true
		}
		_, lineEnd := LineBounds(c.Value, caret)

		return lineEnd, extend, 0, false, true
	default:
		return 0, false, 0, false, false
	}
}

func moveVerticalLogical(value string, pos, dir int) int {
	runes := []rune(value)
	n := len(runes)
	pos = ClampPos(pos, n)
	lineStart, lineEnd := LineBounds(value, pos)
	col := pos - lineStart
	if dir < 0 {
		if lineStart == 0 {
			return 0
		}
		prevEnd := lineStart - 1
		prevStart, _ := LineBounds(value, prevEnd)

		return min(prevStart+col, prevEnd)
	}
	if lineEnd >= n {
		return n
	}
	nextStart := lineEnd + 1
	_, nextEnd := LineBounds(value, nextStart)

	return min(nextStart+col, nextEnd)
}
