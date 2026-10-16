package clipboard

import (
	"context"
	"strings"

	osclip "github.com/chinmay-sawant/ownframe/internal/clipboard"
)

// pasteButton inserts the fixed Go text in the focused field.
func (a *App) pasteButton(ctx context.Context) error {
	if !a.focusLast(ctx) {
		return nil
	}

	if err := a.PasteText(ctx, pasteFromGo); err != nil {
		return err
	}

	a.setStatus("pasted: " + pasteFromGo)

	return nil
}

// pasteClipboard inserts the clipboard text in the last field. The value
// model is one line, so newlines are dropped the way the Ctrl+V chord drops
// them.
func (a *App) pasteClipboard(ctx context.Context) error {
	if !a.focusLast(ctx) {
		a.setStatus("nothing to paste into")

		return nil
	}

	raw, err := osclip.Read()
	if err != nil {
		a.setStatus("clipboard is empty")

		return nil
	}

	text := strings.ReplaceAll(raw, "\r", "")
	text = strings.ReplaceAll(text, "\n", "")
	if text == "" {
		a.setStatus("clipboard is empty")

		return nil
	}

	if err := a.PasteText(ctx, text); err != nil {
		return err
	}

	a.setStatus("pasted: " + text)

	return nil
}
