package ownframe

import "context"

// IMEContext returns the focused field's box and caret text.
// Telegram tests the caret around IME edits with it.
func (p *Page) IMEContext() (Box, int, string, string, bool) {
	return innerPage(p).IMEContext()
}

// IMEReplace replaces a byte range in the focused value.
// Telegram tests IME commits through it.
func (p *Page) IMEReplace(ctx context.Context, start, end int, insert string, caret int) error {
	return innerPage(p).IMEReplace(ctx, start, end, insert, caret)
}
