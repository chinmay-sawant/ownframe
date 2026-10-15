package page

import "context"

// Copy returns the text the page wants on the clipboard.
// A focused text field returns its selected range, or its whole value when
// the range is empty, and does not call the handler.
// ok is false when there is nothing to copy. Copy does not draw.
func (p *Page) Copy(ctx context.Context) (string, bool, error) {
	if err := useContext(ctx); err != nil {
		return "", false, err
	}

	if c, ok := p.focusedEditable(); ok {
		return p.selectedText(c), true, nil
	}

	if p.handlers.Copy == nil {
		return "", false, nil
	}

	return p.handlers.Copy(ctx)
}

// selectedText returns the selected range of c, or the whole value when the
// range is empty.
func (p *Page) selectedText(c Control) string {
	runes := []rune(c.Value)
	start, end := p.formState().bounds(len(runes))
	if start == end {
		return c.Value
	}

	return string(runes[start:end])
}

func useContext(ctx context.Context) error {
	if ctx == nil {
		return errNilContext
	}

	return ctx.Err()
}
