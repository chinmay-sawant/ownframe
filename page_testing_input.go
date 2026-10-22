package ownframe

import "context"

// Click hit-tests the last picture and applies the click.
func (page *Page) Click(ctx context.Context, x, y float64) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Click(ctx, x, y)
}

// Hover sets the hovered element from a point.
func (page *Page) Hover(ctx context.Context, x, y float64) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Hover(ctx, x, y)
}

// Press sets the pressed element from a point.
func (page *Page) Press(ctx context.Context, x, y float64) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Press(ctx, x, y)
}

// Release clears the pressed element.
func (page *Page) Release(ctx context.Context) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Release(ctx)
}

// Type appends text to the focused field.
func (page *Page) Type(ctx context.Context, text string) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Type(ctx, text)
}

// LongPress fires a held press at a point.
func (t Testing) LongPress(ctx context.Context, x, y float64) (bool, error) {
	p := innerPage(t.page)
	if p == nil {
		return false, nil
	}

	return p.LongPress(ctx, x, y)
}

// Swipe delivers a swipe delta to the handler.
func (t Testing) Swipe(ctx context.Context, dx, dy float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Swipe(ctx, dx, dy)
}

// BeginDrag starts a pointer drag.
func (t Testing) BeginDrag(ctx context.Context, x, y float64) (bool, error) {
	p := innerPage(t.page)
	if p == nil {
		return false, nil
	}

	return p.BeginDrag(ctx, x, y)
}

// MoveDrag continues a pointer drag.
func (t Testing) MoveDrag(ctx context.Context, x, y float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.MoveDrag(ctx, x, y)
}

// EndDrag finishes a pointer drag.
func (t Testing) EndDrag(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.EndDrag(ctx)
}

// Backspace drops the last rune of the focused field.
func (page *Page) Backspace(ctx context.Context) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Backspace(ctx)
}

// DeleteWord deletes the word behind the caret.
func (page *Page) DeleteWord(ctx context.Context) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.DeleteWord(ctx)
}

// Submit calls the submit handler and draws.
func (page *Page) Submit(ctx context.Context) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Submit(ctx)
}

// KeyDown calls the key-down handler and caret keys.
func (page *Page) KeyDown(ctx context.Context, key string) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.KeyDown(ctx, key)
}

// KeyUp calls the key-up handler.
func (page *Page) KeyUp(ctx context.Context, key string) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.KeyUp(ctx, key)
}

// CancelTouches queues gesture cancellation for the window loop.
func (page *Page) CancelTouches() {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.CancelTouches()
}

// Drop calls the drop handler with the files dropped on the window.
func (page *Page) Drop(ctx context.Context, files []Drop) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Drop(ctx, files)
}

// IMEReplace replaces a byte range in the focused value.
func (t Testing) IMEReplace(ctx context.Context, start, end int, insert string, caret int) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.IMEReplace(ctx, start, end, insert, caret)
}

// IMEContext returns the focused field's box and caret text.
func (t Testing) IMEContext() (Box, int, string, string, bool) {
	p := innerPage(t.page)
	if p == nil {
		return Box{}, 0, "", "", false
	}

	return p.IMEContext()
}

// Copy returns the text the page wants on the clipboard.
func (page *Page) Copy(ctx context.Context) (string, bool, error) {
	p := innerPage(page)
	if p == nil {
		return "", false, nil
	}

	return p.Copy(ctx)
}

// Cut returns the selected text and removes it.
func (page *Page) Cut(ctx context.Context) (string, bool, error) {
	p := innerPage(page)
	if p == nil {
		return "", false, nil
	}

	return p.Cut(ctx)
}

// Paste inserts text into the focused field.
func (page *Page) Paste(ctx context.Context, text string) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Paste(ctx, text)
}

// SelectAll selects the whole value of a focused text field.
func (page *Page) SelectAll(ctx context.Context) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.SelectAll(ctx)
}

// Undo calls the undo handler and draws.
func (page *Page) Undo(ctx context.Context) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Undo(ctx)
}

// Redo calls the redo handler and draws.
func (page *Page) Redo(ctx context.Context) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Redo(ctx)
}

// SelectAt focuses the control under the point and places the caret.
func (t Testing) SelectAt(ctx context.Context, x, y float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.SelectAt(ctx, x, y)
}

// LastFocusID is the last focused control id.
func (page *Page) LastFocusID() string {
	p := innerPage(page)
	if p == nil {
		return ""
	}

	return p.LastFocusID()
}

// FocusedField returns the focused field id, excluding buttons.
func (page *Page) FocusedField() string {
	p := innerPage(page)
	if p == nil {
		return ""
	}

	return p.FocusedField()
}
