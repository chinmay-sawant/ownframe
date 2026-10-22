package page

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

// caretBlinkInterval is how long the caret stays shown or hidden.
const caretBlinkInterval = form.BlinkInterval

// resetCaretBlink shows the caret and restarts the blink clock. Every caret
// move and edit calls it, so the caret is solid while the person types.
func (p *Page) resetCaretBlink() {
	if p.formState() == nil {
		return
	}

	p.formState().blinkOn, p.formState().blinkAt = form.ResetBlink()
}

// caretBlink toggles the focused field's caret on the blink clock and draws
// when the phase changes. A range selection paints no caret, so it does
// nothing. A file input never shows a caret, so it does nothing.
func (p *Page) caretBlink(ctx context.Context) error {
	c, ok := p.focusedEditable()
	if !ok || isFileInput(c) {
		return nil
	}

	n := runeLen(c.Value)
	start, end := p.formState().bounds(n)
	if start != end {
		return nil
	}

	on, at, draw := form.AdvanceBlink(p.formState().blinkOn, p.formState().blinkAt, time.Now())
	p.formState().blinkOn, p.formState().blinkAt = on, at
	if !draw {
		return nil
	}

	return p.Redraw(ctx)
}
