package ownframe

import "context"

// Backspace drops the last rune of the focused field.
func (t Testing) Backspace(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Backspace(ctx)
}

// DeleteWord deletes the word behind the caret.
func (t Testing) DeleteWord(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.DeleteWord(ctx)
}

// Submit calls the submit handler and draws.
func (t Testing) Submit(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Submit(ctx)
}

// KeyDown calls the key-down handler and caret keys.
func (t Testing) KeyDown(ctx context.Context, key string) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.KeyDown(ctx, key)
}

// KeyUp calls the key-up handler.
func (t Testing) KeyUp(ctx context.Context, key string) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.KeyUp(ctx, key)
}
