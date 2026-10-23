package gesture

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
	"github.com/chinmay-sawant/ownframe/internal/textrun"
)

// Preedit is temporary IME text. ID is the focused field it belongs to.
// Start and End are byte offsets within Text.
type Preedit struct {
	ID         string
	Text       string
	Start, End int
}

// SplitCaret splits text at a rune caret. The int is the byte length of
// the text before the caret.
func SplitCaret(text string, caret int) (before, after string, bytes int) {
	runes := []rune(text)
	caret = form.ClampPos(caret, len(runes))
	before = string(runes[:caret])

	return before, string(runes[caret:]), len(before)
}

// Unscroll moves box by the page scroll offset, in CSS pixels.
func Unscroll(box layout.Box, scrollX, scrollY float64) layout.Box {
	box.X -= scrollX
	box.Y -= scrollY

	return box
}

// Compose decides the next preedit. redraw is false when the call is a no-op:
// an empty clear, a field that cannot take composition, or the same range.
// Clearing a visible preedit returns the zero value.
func Compose(cur Preedit, focusID, text string, start, end int, hasField bool) (Preedit, bool) {
	if text == "" {
		if cur.Text == "" {
			return cur, false
		}

		return Preedit{}, true
	}

	if !hasField {
		return cur, false
	}

	same := cur.ID == focusID && cur.Text == text && cur.Start == start && cur.End == end
	if same {
		return cur, false
	}

	return Preedit{ID: focusID, Text: text, Start: start, End: end}, true
}

// CaretSample is the text and rune offset the caret anchor measures.
// text is the field value. A preedit for another field is ignored.
func CaretSample(field form.Control, text string, caret, anchor int, pre Preedit) (string, int) {
	kind := strings.ToLower(strings.TrimSpace(field.Type))
	if pre.ID != field.ID {
		return form.ShownText(kind, text), form.ClampPos(caret, form.RuneLen(text))
	}

	base := []rune(text)
	start, end := form.RangeOf(anchor, caret, len(base))
	value := string(base[:start]) + pre.Text + string(base[end:])
	offset := start + form.RuneAtByte(pre.Text, pre.Start)

	return form.ShownText(kind, value), offset
}

// Replace replaces bytes [start, end) of the field value with text and
// returns the new value and the caret rune. caret is a byte offset within
// text. changed is false when the call would delete an empty range.
func Replace(field form.Control, text string, start, end, caret int) (string, int, bool) {
	value := field.Value
	runes := []rune(value)
	startRune := form.RuneAtByte(value, start)
	endRune := form.RuneAtByte(value, end)
	if startRune == endRune && text == "" {
		return value, 0, false
	}

	ins := []rune(text)
	pos := startRune + form.ClampPos(form.RuneAtByte(text, caret), len(ins))
	out := make([]rune, 0, len(runes)-(endRune-startRune)+len(ins))
	out = append(out, runes[:startRune]...)
	out = append(out, ins...)
	out = append(out, runes[endRune:]...)

	return string(out), pos, true
}

// Anchor is the focused glyph caret in CSS pixels, with scroll removed.
// ok is false when the display cannot measure the caret.
func Anchor(d *layout.Display, field form.Control, text string, caret, anchor int, pre Preedit, scrollX, scrollY float64) (layout.Box, bool) {
	if d == nil {
		return layout.Box{}, false
	}

	box, ok := form.CaretBox(d, field.ID)
	if !ok {
		return layout.Box{}, false
	}

	sampleText, offset := CaretSample(field, text, caret, anchor, pre)
	sample, x, ok := textrun.Caret(d, box, sampleText, offset)
	if !ok {
		sample, x = form.EmptyCaret(d, box)
	}

	pt := d.PixelPerPoint

	return layout.Box{
		X: x*pt - scrollX,
		Y: (sample.Y-sample.Size)*pt - scrollY,
		W: d.PointsPerPixel * pt,
		H: sample.Size * pt,
	}, true
}
