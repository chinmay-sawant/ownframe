package page

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/textrun"
)

func (p *Page) paintCaret(d *layout.Display) {
	if p.form == nil || p.form.focusID == "" || p.form.all {
		return
	}
	c, ok := p.focusedEditable()
	if !ok || isFileInput(c) {
		return
	}
	start, end := p.form.bounds(runeLen(c.Value))
	if start != end {
		return
	}
	text, offset, ok := p.caretText(c)
	if !ok {
		return
	}
	box, ok := caretBox(d, c.ID)
	if !ok {
		return
	}
	sample, x, ok := textrun.Caret(d, box, text, offset)
	if !ok {
		sample, x = emptyCaret(d, box)
	}
	if sample.Size <= 0 {
		sample.Size = 12
	}
	sample.Kind = layout.DisplayOpNoop
	if p.form.blinkOn {
		sample.Kind = layout.DisplayOpFillRect
	}
	sample.X, sample.Y = x, sample.Y-sample.Size
	sample.W, sample.H = d.PointsPerPixel, sample.Size
	sample.Text, sample.Font = "", nil
	d.Ops = append(d.Ops, sample)
	d.Order = append(d.Order, len(d.Ops)-1)
}

func caretBox(d *layout.Display, id string) (layout.Box, bool) {
	for _, box := range d.Boxes {
		if box.ID == id {
			return box, true
		}
	}
	return layout.Box{}, false
}

func emptyCaret(d *layout.Display, box layout.Box) (layout.DisplayOp, float64) {
	pt := d.PointsPerPixel
	return layout.DisplayOp{Y: (box.Y + box.H*.65) * pt, Size: 12, Alpha: 1}, (box.X + 7) * pt
}

func (p *Page) caretText(c Control) (string, int, bool) {
	kind := strings.ToLower(strings.TrimSpace(c.Type))
	if p.form.preeditID != c.ID {
		return shownText(kind, c.Value), clampPos(p.form.caret, runeLen(c.Value)), true
	}
	if p.form.preeditStart != p.form.preeditEnd {
		return "", 0, false
	}
	base := []rune(c.Value)
	start, end := p.form.bounds(len(base))
	value := string(base[:start]) + p.form.preedit + string(base[end:])
	return shownText(kind, value), start + runeAtByte(p.form.preedit, p.form.preeditStart), true
}
