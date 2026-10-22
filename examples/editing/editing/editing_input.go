package editing

import "context"

// Type edits the focused note.
func (a *App) Type(ctx context.Context, text string) error {
	return a.page.Type(ctx, text)
}

// Backspace drops the last rune of the focused note.
func (a *App) Backspace(ctx context.Context) error {
	return a.page.Backspace(ctx)
}

// Paste inserts text into the focused note.
func (a *App) Paste(ctx context.Context, text string) error {
	return a.page.Paste(ctx, text)
}

// Cut returns the focused note and clears it.
func (a *App) Cut(ctx context.Context) (string, bool, error) {
	return a.page.Cut(ctx)
}

// Copy returns the focused note without changing it.
func (a *App) Copy(ctx context.Context) (string, bool, error) {
	return a.page.Copy(ctx)
}
