package telegram

import "context"

// RequestBlur queues keyboard dismissal on the game loop.
func (a *App) RequestBlur() {
	a.mu.Lock()
	a.blur = true
	a.mu.Unlock()
}

func (a *App) applyBlur(ctx context.Context) error {
	a.mu.Lock()
	blur := a.blur
	a.blur = false
	a.mu.Unlock()
	if blur {
		return a.page.Focus(ctx, "")
	}
	return nil
}
