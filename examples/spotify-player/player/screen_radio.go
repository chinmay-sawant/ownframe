package player

import (
	"context"
	"strings"
)

// clickRadio applies one action on the radio screen.
func (a *App) clickRadio(ctx context.Context, action string) error {
	switch {
	case action == "radio-play":
		a.RadioTogglePlay(ctx)
	case strings.HasPrefix(action, "radio-st-"):
		a.RadioSelectStation(ctx, slot(action, "radio-st-"))
	}

	return nil
}

// RadioTogglePlay mirrors the now-bar play button for the radio screen.
func (a *App) RadioTogglePlay(ctx context.Context) {
	a.view.Playing = !a.view.Playing

	if a.view.Playing {
		if a.audio.Duration() > 0 {
			a.audio.Play()
		} else {
			a.playNow(ctx)
		}

		return
	}

	a.audio.Pause()
}

// RadioSelectStation makes station i the now-playing item and starts it.
func (a *App) RadioSelectStation(ctx context.Context, i int) {
	if i < 0 || i >= len(a.view.Radio.Stations) {
		return
	}

	for j := range a.view.Radio.Stations {
		a.view.Radio.Stations[j].Active = j == i
	}

	a.setNow(a.view.Radio.Stations[i])
	a.playNow(ctx)
}
