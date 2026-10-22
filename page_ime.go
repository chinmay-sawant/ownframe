package ownframe

// IMEContext returns the focused field's box and caret text.
// Telegram tests the caret around IME edits with it.
func (p *Page) IMEContext() (Box, int, string, string, bool) {
	return innerPage(p).IMEContext()
}
