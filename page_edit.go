package ownframe

import "context"

// Copy returns the text the page wants on the clipboard.
func (p *Page) Copy(ctx context.Context) (string, bool, error) {
	return innerPage(p).Copy(ctx)
}

// Cut returns the selected text and removes it.
func (p *Page) Cut(ctx context.Context) (string, bool, error) {
	return innerPage(p).Cut(ctx)
}

// Paste inserts text into the focused field.
func (p *Page) Paste(ctx context.Context, text string) error {
	return innerPage(p).Paste(ctx, text)
}

// SelectAll selects the whole value of a focused text field.
func (p *Page) SelectAll(ctx context.Context) error {
	return innerPage(p).SelectAll(ctx)
}

// Undo calls the undo handler and draws.
func (p *Page) Undo(ctx context.Context) error {
	return innerPage(p).Undo(ctx)
}

// Redo calls the redo handler and draws.
func (p *Page) Redo(ctx context.Context) error {
	return innerPage(p).Redo(ctx)
}

// SelectAt focuses the control under the point and places the caret.
func (p *Page) SelectAt(ctx context.Context, x, y float64) error {
	return innerPage(p).SelectAt(ctx, x, y)
}
