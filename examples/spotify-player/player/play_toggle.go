package player

import (
	"context"
	"time"
)

// setPlaying flips playback and arms the virtual clock on start so the bar
// moves on the first frame even before a frozen engine reports motion. A
// live engine disengages it on its first move. Starting from a stopped
// position resets the bar first, so the committed template, the retained
// ops, and the virtual clock agree on zero instead of a stale default.
func (a *App) setPlaying(ctx context.Context, play bool) {
	a.view.Playing = play
	if !a.view.Playing {
		a.audio.Pause()
		return
	}
	pos := a.audio.Position()
	a.vpos = pos
	a.vlastPos = pos
	a.vlast = time.Now()
	a.vfreeze = 0
	a.vengaged = true
	if pos == 0 {
		a.resetProgress()
	}
	if a.audio.Duration() > 0 {
		a.audio.Play()
		return
	}
	a.playNow(ctx)
}

// togglePlayback starts or pauses the current audio item.
func (a *App) togglePlayback(ctx context.Context) {
	a.setPlaying(ctx, !a.view.Playing)
}
