package page

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
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
		sample.Size = caretFontSize
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

func (p *Page) caretText(c Control) (string, int, bool) {
	fs := p.formState()

	return form.CaretText(form.FocusView{
		Caret:        fs.caret,
		Anchor:       fs.anchor,
		Control:      c,
		PreeditID:    fs.preeditID,
		Preedit:      fs.preedit,
		PreeditStart: fs.preeditStart,
		PreeditEnd:   fs.preeditEnd,
	})
}
