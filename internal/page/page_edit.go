package page

import "context"

// Paste calls the paste handler and edits a focused text field.
// It draws when the handler ran or the field changed.
// A nil handler with nothing to edit does nothing.
func (p *Page) Paste(ctx context.Context, text string) error {
	var fn func() error
	if p.handlers.Paste != nil {
		fn = func() error {
			return p.handlers.Paste(ctx, text)
		}
	}

	return p.editField(ctx, fn, func(c Control, caret, anchor int) (Control, int, int, bool) {
		return p.insertValue(c, caret, anchor, text)
	})
}

// SelectAll selects the whole value of a focused text field, or calls the
// select-all handler. A focused field does not call the handler. It draws
// after selecting. It does nothing when no field is focused and no handler
// is registered.
func (p *Page) SelectAll(ctx context.Context) error {
	if err := useContext(ctx); err != nil {
		return err
	}

	if c, ok := p.typingTarget(); ok {
		n := runeLen(c.Value)
		p.formState().anchor, p.formState().caret = 0, n
		p.formState().all = true
		p.resetCaretBlink()

		return p.Redraw(ctx)
	}

	return p.after(ctx, p.handlers.SelectAll == nil, func() error {
		return p.handlers.SelectAll(ctx)
	})
}

// Undo calls the undo handler and draws the page again.
// It does nothing when no undo handler is registered.
func (p *Page) Undo(ctx context.Context) error {
	return p.after(ctx, p.handlers.Undo == nil, func() error {
		return p.handlers.Undo(ctx)
	})
}

// Redo calls the redo handler and draws the page again.
// It does nothing when no redo handler is registered.
func (p *Page) Redo(ctx context.Context) error {
	return p.after(ctx, p.handlers.Redo == nil, func() error {
		return p.handlers.Redo(ctx)
	})
}
