package editing

import "context"

// Type edits the focused note.
func (a *App) Type(ctx context.Context, text string) error {
	return a.page.Edit().Type(ctx, text)
}

// Backspace drops the last rune of the focused note.
func (a *App) Backspace(ctx context.Context) error {
	return a.page.Edit().Backspace(ctx)
}

// Paste inserts text into the focused note.
func (a *App) Paste(ctx context.Context, text string) error {
	return a.page.Edit().Paste(ctx, text)
}

// Cut returns the focused note and clears it.
func (a *App) Cut(ctx context.Context) (string, bool, error) {
	return a.page.Edit().Cut(ctx)
}

// Copy returns the focused note without changing it.
func (a *App) Copy(ctx context.Context) (string, bool, error) {
	return a.page.Edit().Copy(ctx)
}
