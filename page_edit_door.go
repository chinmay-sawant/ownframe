package ownframe

import "context"

// Edit runs text edits, keys, and pointer actions.
type Edit struct{ page *Page }

// Edit returns text edits, keys, and pointer actions for this page.
func (p *Page) Edit() Edit { return Edit{page: p} }

// Copy returns the text the page wants on the clipboard.
func (e Edit) Copy(ctx context.Context) (string, bool, error) {
	p := innerPage(e.page)
	if p == nil {
		return "", false, nil
	}

	return p.Copy(ctx)
}

// Cut returns the selected text and removes it.
func (e Edit) Cut(ctx context.Context) (string, bool, error) {
	p := innerPage(e.page)
	if p == nil {
		return "", false, nil
	}

	return p.Cut(ctx)
}

// Paste inserts text into the focused field.
func (e Edit) Paste(ctx context.Context, text string) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.Paste(ctx, text)
}

// SelectAll selects the whole value of a focused text field.
func (e Edit) SelectAll(ctx context.Context) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.SelectAll(ctx)
}

// Undo calls the undo handler and draws.
func (e Edit) Undo(ctx context.Context) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.Undo(ctx)
}

// Redo calls the redo handler and draws.
func (e Edit) Redo(ctx context.Context) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.Redo(ctx)
}

// Type appends text to the focused field.
func (e Edit) Type(ctx context.Context, text string) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.Type(ctx, text)
}

// Backspace drops the last rune of the focused field.
func (e Edit) Backspace(ctx context.Context) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.Backspace(ctx)
}

// DeleteWord deletes the word behind the caret.
func (e Edit) DeleteWord(ctx context.Context) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.DeleteWord(ctx)
}

// Submit calls the submit handler and draws.
func (e Edit) Submit(ctx context.Context) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.Submit(ctx)
}

// KeyDown calls the key-down handler and caret keys.
func (e Edit) KeyDown(ctx context.Context, key string) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.KeyDown(ctx, key)
}

// KeyUp calls the key-up handler.
func (e Edit) KeyUp(ctx context.Context, key string) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.KeyUp(ctx, key)
}

// Hover sets the hovered element from a point.
func (e Edit) Hover(ctx context.Context, x, y float64) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.Hover(ctx, x, y)
}

// Press sets the pressed element from a point.
func (e Edit) Press(ctx context.Context, x, y float64) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.Press(ctx, x, y)
}

// Release clears the pressed element.
func (e Edit) Release(ctx context.Context) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.Release(ctx)
}

// Drop calls the drop handler with the files dropped on the window.
func (e Edit) Drop(ctx context.Context, files []Drop) error {
	p := innerPage(e.page)
	if p == nil {
		return nil
	}

	return p.Drop(ctx, files)
}

// CancelTouches queues gesture cancellation for the window loop.
func (e Edit) CancelTouches() {
	p := innerPage(e.page)
	if p == nil {
		return
	}
	p.CancelTouches()
}

// LastFocusID is the last focused control id.
func (e Edit) LastFocusID() string {
	p := innerPage(e.page)
	if p == nil {
		return ""
	}

	return p.LastFocusID()
}

// FocusedField returns the focused field id, excluding buttons.
func (e Edit) FocusedField() string {
	p := innerPage(e.page)
	if p == nil {
		return ""
	}

	return p.FocusedField()
}
