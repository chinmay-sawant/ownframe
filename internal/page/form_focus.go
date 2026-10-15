package page

import "context"

// FocusNext focuses the next control in document order and wraps. Disabled
// controls and tabindex="-1" are skipped. It draws when the focus moved.
func (p *Page) FocusNext(ctx context.Context) error {
	return p.focusStep(ctx, 1)
}

// FocusPrev focuses the previous control in document order and wraps.
// Disabled controls and tabindex="-1" are skipped. It draws when the focus
// moved.
func (p *Page) FocusPrev(ctx context.Context) error {
	return p.focusStep(ctx, -1)
}

// Focus focuses the control id and draws. An empty id blurs. A disabled
// control or an unknown id is left alone.
func (p *Page) Focus(ctx context.Context, id string) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	if id == "" {
		if p.formState() == nil || p.formState().focusID == "" {
			return nil
		}

		p.blurForm()

		return p.Redraw(ctx)
	}

	c, ok := p.control(id)
	if !ok || c.Disabled {
		return nil
	}

	if p.formState().focusID == id {
		return nil
	}

	p.formState().focusID = id
	if c.Tag == "button" {
		p.revealFocus(id)
	}
	if canEdit(c) {
		p.caretEnd(c)
	} else {
		p.clearRange()
	}

	return p.Redraw(ctx)
}

// FocusID returns the focused control id, or "" when none is focused.
func (p *Page) FocusID() string {
	if p.formState() == nil {
		return ""
	}

	return p.formState().focusID
}

func (p *Page) focusStep(ctx context.Context, step int) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	order := p.focusOrder()
	if len(order) == 0 {
		return nil
	}

	at := -1
	for i, id := range order {
		if id == p.FocusID() {
			at = i

			break
		}
	}

	next := 0
	if at < 0 {
		if step < 0 {
			next = len(order) - 1
		}
	} else {
		next = (at + step + len(order)) % len(order)
	}

	return p.Focus(ctx, order[next])
}
