package ownframe

import "context"

// Click hit-tests the last picture and applies the click.
func (p *Page) Click(ctx context.Context, x, y float64) error {
	return innerPage(p).Click(ctx, x, y)
}

// Hover sets the hovered element from a point.
func (p *Page) Hover(ctx context.Context, x, y float64) error {
	return innerPage(p).Hover(ctx, x, y)
}

// Press sets the pressed element from a point.
func (p *Page) Press(ctx context.Context, x, y float64) error {
	return innerPage(p).Press(ctx, x, y)
}

// Release clears the pressed element.
func (p *Page) Release(ctx context.Context) error {
	return innerPage(p).Release(ctx)
}

// Type appends text to the focused field.
func (p *Page) Type(ctx context.Context, text string) error {
	return innerPage(p).Type(ctx, text)
}

// Backspace drops the last rune of the focused field.
func (p *Page) Backspace(ctx context.Context) error {
	return innerPage(p).Backspace(ctx)
}

// DeleteWord deletes the word behind the caret.
func (p *Page) DeleteWord(ctx context.Context) error {
	return innerPage(p).DeleteWord(ctx)
}

// Submit calls the submit handler and draws.
func (p *Page) Submit(ctx context.Context) error {
	return innerPage(p).Submit(ctx)
}

// KeyDown calls the key-down handler and caret keys.
func (p *Page) KeyDown(ctx context.Context, key string) error {
	return innerPage(p).KeyDown(ctx, key)
}

// KeyUp calls the key-up handler.
func (p *Page) KeyUp(ctx context.Context, key string) error {
	return innerPage(p).KeyUp(ctx, key)
}
