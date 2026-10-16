package page

import (
	"context"
	"time"
)

// caretBlinkInterval is how long the caret stays shown or hidden.
const caretBlinkInterval = 530 * time.Millisecond

// resetCaretBlink shows the caret and restarts the blink clock. Every caret
// move and edit calls it, so the caret is solid while the person types.
func (p *Page) resetCaretBlink() {
	if p.formState() == nil {
		return
	}

	p.formState().blinkOn = true
	p.formState().blinkAt = time.Time{}
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

	now := time.Now()
	if p.formState().blinkAt.IsZero() {
		p.formState().blinkAt = now
		p.formState().blinkOn = true

		return nil
	}

	if now.Sub(p.formState().blinkAt) < caretBlinkInterval {
		return nil
	}

	p.formState().blinkOn = !p.formState().blinkOn
	p.formState().blinkAt = now

	return p.Redraw(ctx)
}
