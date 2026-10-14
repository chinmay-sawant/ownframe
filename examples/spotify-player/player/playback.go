package player

import (
	"context"
	"math/rand"
	"strings"
)

// Close releases the audio device. main defers it after a successful New.
func (a *App) Close() {
	if a.audio != nil {
		a.audio.Close()
	}
	a.closeLog()
}

// Tick advances audio and animation once per frame.
func (a *App) Tick(ctx context.Context) error {
	a.page.UseFrameDirty()
	a.audio.Update()
	a.autoStart(ctx)

	if !a.audio.Loading() {
		if err := a.advance(ctx); err != nil {
			return err
		}
	}

	a.animate()

	return nil
}

// requestKey names the audio request behind the current now-playing item.
func (a *App) requestKey() string {
	return strings.TrimSpace(a.view.Now.Title) + "|" + strings.TrimSpace(a.view.Now.Artist)
}

// playNow asks the engine for the bundled clip at the current row index.
func (a *App) playNow(ctx context.Context) {
	query := strings.TrimSpace(a.view.Now.Genre)
	if query == "" {
		query = "instrumental"
	}

	a.audioKey = a.requestKey()
	a.audio.Want(ctx, query, a.view.Now.Index)
	a.view.Playing = true
}

// autoStart begins the current track when playing is wanted but the engine
// has not been asked for this request key yet.
func (a *App) autoStart(ctx context.Context) {
	if !a.view.Playing || a.audio.Loading() || a.audio.Err() != nil {
		return
	}

	if a.audioKey == a.requestKey() {
		return
	}

	a.playNow(ctx)
}

// advance moves to the next track when the voice finishes.
func (a *App) advance(ctx context.Context) error {
	if !a.view.Playing || !a.audio.Ended() {
		return nil
	}

	if a.view.Repeat {
		a.audio.SeekFraction(0)
		a.audio.Play()

		return nil
	}

	if a.view.Shuffle && len(a.view.Tracks) > 0 {
		a.selectRow(rand.Intn(len(a.view.Tracks)))
	} else {
		a.step(1)
	}

	a.playNow(ctx)
	a.page.SetData(a.view)

	return a.page.Redraw(ctx)
}
