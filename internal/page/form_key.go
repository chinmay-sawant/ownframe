package page

import "context"

// valueEdit computes one edit to a field: the new value, the caret and the
// anchor to store, and whether anything changed.
type valueEdit func(c Control, caret, anchor int) (Control, int, int, bool)

func (p *Page) editField(ctx context.Context, fn func() error, edit valueEdit) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	ran := false
	if fn != nil {
		if err := fn(); err != nil {
			return err
		}

		ran = true
	}

	changed, err := p.applyEdit(ctx, edit)
	if err != nil {
		return err
	}

	if !changed && !ran {
		return nil
	}

	if changed {
		p.markPending(p.formState().focusID)

		if err := p.change(ctx, p.formState().focusID); err != nil {
			return err
		}
	}

	return p.Redraw(ctx)
}

func (p *Page) applyEdit(ctx context.Context, edit valueEdit) (bool, error) {
	c, ok := p.typingTarget()
	if !ok {
		return false, nil
	}

	next, caret, anchor, changed := edit(c, p.formState().caret, p.formState().anchor)
	if !changed {
		return false, nil
	}

	// GEN-28: every edit funnels here, so clamp once for all writers.
	next.Value = clampField(next.Value, fieldCap(next))
	if n := runeLen(next.Value); caret > n || anchor > n {
		caret, anchor = min(caret, n), min(anchor, n)
	}

	if err := p.beforeEdit(ctx, p.formState().focusID); err != nil {
		return false, err
	}

	p.formState().byID[p.formState().focusID] = next
	p.formState().caret, p.formState().anchor = caret, anchor
	p.formState().all = false
	p.resetCaretBlink()
	bindWrite(p, next)

	return true, nil
}

func (p *Page) typingTarget() (Control, bool) {
	c, ok := p.focusedEditable()
	if !ok || c.Disabled {
		return Control{}, false
	}

	return c, true
}

func (p *Page) focusedEditable() (Control, bool) {
	if p.formState() == nil {
		return Control{}, false
	}

	c, ok := p.control(p.formState().focusID)
	if !ok || !canEdit(c) {
		return Control{}, false
	}

	return c, true
}
