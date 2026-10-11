package host

import (
	"context"

	"github.com/chinmay-sawant/blinkless/layout"
)

// IME is a screen that takes platform text input, such as the Android or
// iOS soft keyboard. The window drives one session for the focused text
// control and hands committed text back to the screen.
type IME interface {
	// IMEContext returns the focused text control's display box, its caret
	// as a byte offset into the value, and the value around the caret. ok is
	// false when no text control is focused.
	IMEContext() (box layout.Box, caret int, before, after string, ok bool)

	// IMEReplace replaces the focused value's bytes [start, end) with
	// insert and puts the caret at the byte offset caret within insert. It
	// draws again.
	IMEReplace(ctx context.Context, start, end int, insert string, caret int) error
}

// IMEPreeditor receives temporary composition text without changing the
// committed control value.
type IMEPreeditor interface {
	SetIMEComposition(ctx context.Context, text string, start, end int) error
}
