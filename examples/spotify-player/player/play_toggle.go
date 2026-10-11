package player

import "context"

// togglePlayback starts or pauses the current audio item.
func (a *App) togglePlayback(ctx context.Context) {
	a.view.Playing = !a.view.Playing
	if !a.view.Playing {
		a.audio.Pause()
		return
	}
	if a.audio.Duration() > 0 {
		a.audio.Play()
		return
	}
	a.playNow(ctx)
}
