package page

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe/internal/textrun"
)

// offsetAt turns a point into a rune offset in c's value. A run that cannot
// be measured, or one whose drawn text does not line up with the shown
// value, lands at the end of the value.
func (p *Page) offsetAt(box Box, c Control, x, y float64) int {
	n := runeLen(c.Value)
	off, ok := textrun.At(p.display, box, x, y, shownText(strings.ToLower(c.Type), c.Value))
	if !ok {
		return n
	}

	return clampPos(off, n)
}

// focusCaret focuses c and puts the caret and the anchor at off.
func (p *Page) focusCaret(ctx context.Context, c Control, off int) error {
	changed := p.formState().focusID != c.ID || p.formState().caret != off ||
		p.formState().anchor != off || p.formState().all
	p.formState().focusID = c.ID
	p.formState().caret, p.formState().anchor = off, off
	p.formState().all = false
	p.resetCaretBlink()
	if !changed {
		return nil
	}

	return p.Redraw(ctx)
}
