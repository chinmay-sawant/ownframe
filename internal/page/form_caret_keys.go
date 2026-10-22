package page

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

// caretKey moves the caret for a focused text field and draws. Arrow keys
// move by rune, Home and End move to the ends, and a ctrl+ or alt+ prefix
// jumps by word. A shift+ prefix extends the range from the anchor.
func (p *Page) caretKey(ctx context.Context, key string) error {
	c, ok := p.typingTarget()
	if !ok {
		return nil
	}

	n := runeLen(c.Value)
	caret := clampPos(p.formState().caret, n)
	start, end := p.formState().bounds(n)
	pos, extend, ok := form.MoveCaret(key, c.Value, caret, start, end)
	if !ok {
		return nil
	}

	changed := pos != p.formState().caret
	if !extend {
		changed = changed || p.formState().anchor != pos || p.formState().all
	}

	if !changed {
		return nil
	}

	p.setCaret(pos, extend)

	return p.Redraw(ctx)
}
