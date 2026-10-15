package page

import (
	"strings"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/textrun"
)

func (p *Page) paintCaret(d *layout.Display) {
	if p.formState() == nil || p.formState().focusID == "" || p.formState().all {
		return
	}
	c, ok := p.focusedEditable()
	if !ok || isFileInput(c) {
		return
	}
	start, end := p.formState().bounds(runeLen(c.Value))
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
	if p.formState().blinkOn {
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
	if p.formState().preeditID != c.ID {
		return shownText(kind, c.Value), clampPos(p.formState().caret, runeLen(c.Value)), true
	}
	if p.formState().preeditStart != p.formState().preeditEnd {
		return "", 0, false
	}
	base := []rune(c.Value)
	start, end := p.formState().bounds(len(base))
	value := string(base[:start]) + p.formState().preedit + string(base[end:])
	return shownText(kind, value), start + runeAtByte(p.formState().preedit, p.formState().preeditStart), true
}
