package page

import "context"

// Cut returns the text the page wants on the clipboard and draws again.
// A focused text field returns its selected range, or its whole value when
// the range is empty, removes that text, writes the bound field, and fires
// Change before the redraw. It does not call the handler.
// ok is false when there is nothing to cut.
func (p *Page) Cut(ctx context.Context) (string, bool, error) {
	if err := useContext(ctx); err != nil {
		return "", false, err
	}

	if c, ok := p.focusedEditable(); ok {
		id := p.formState().focusID
		if err := p.beforeEdit(ctx, id); err != nil {
			return "", false, err
		}

		text := p.selectedText(c)
		runes := []rune(c.Value)
		start, end := p.formState().bounds(len(runes))
		if start == end {
			start, end = 0, len(runes)
		}

		c.Value = string(append(runes[:start], runes[end:]...))
		p.formState().caret, p.formState().anchor = start, start
		p.formState().all = false
		p.resetCaretBlink()
		p.formState().byID[id] = c
		p.markPending(id)
		bindWrite(p, c)
		if err := p.change(ctx, id); err != nil {
			return text, true, err
		}
		err := p.Redraw(ctx)

		return text, true, err
	}

	if p.handlers.Cut == nil {
		return "", false, nil
	}

	text, ok, err := p.handlers.Cut(ctx)
	if err != nil || !ok {
		return text, ok, err
	}

	if err := p.Redraw(ctx); err != nil {
		return text, ok, err
	}

	return text, ok, nil
}
