package host

import "context"

// Pointer is the mouse and touch surface of a screen. Window pointer,
// touch, and drag code takes this instead of the full Screen so fakes
// cover four methods instead of the whole page surface.
type Pointer interface {
	Click(ctx context.Context, x, y float64) error
	Hover(ctx context.Context, x, y float64) error
	Press(ctx context.Context, x, y float64) error
	Release(ctx context.Context) error
}

// Keyboard is the typing surface of a screen. Key, text, and submit
// code takes this so fakes skip pointer and clipboard methods.
type Keyboard interface {
	Type(ctx context.Context, text string) error
	KeyDown(ctx context.Context, key string) error
	KeyUp(ctx context.Context, key string) error
	Backspace(ctx context.Context) error
	DeleteWord(ctx context.Context) error
	Submit(ctx context.Context) error
}

// Clipboard is the edit surface around the OS clipboard. Chord and
// menu code takes this; call logic is unchanged, only the asserted
// type narrows.
type Clipboard interface {
	Copy(ctx context.Context) (string, bool, error)
	Cut(ctx context.Context) (string, bool, error)
	Paste(ctx context.Context, text string) error
	SelectAll(ctx context.Context) error
	Undo(ctx context.Context) error
	Redo(ctx context.Context) error
}
