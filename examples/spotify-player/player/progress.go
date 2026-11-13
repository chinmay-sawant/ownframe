package player

// cardLength is the duration a card plays when it has no track length.
const cardLength = "3:30"

// setNow swaps the now-playing item for a card.
func (a *App) setNow(card Card) {
	a.view.Now = Track{Index: card.Index, Title: clip(card.Title, trackTitleClip), Artist: clip(card.Sub, trackTitleClip), Length: cardLength, Cover: card.Cover}
	a.resetProgress()
}

// toggleMute drops the volume to zero, or restores the remembered level.
func (a *App) toggleMute() {
	if a.view.Volume == 0 {
		a.view.Volume = a.lastVolume
		if a.view.Volume == 0 {
			a.view.Volume = defaultVolume
		}

		return
	}

	a.lastVolume = a.view.Volume
	a.view.Volume = 0
}

// syncViewFromAudio refreshes Progress/Elapsed/Remaining/Credit from the
// engine. Call only before a real commit (SetData plus Redraw), never from
// the tick path: the tick paints the same values through retained ops.
func (a *App) syncViewFromAudio() {
	pos, dur := a.audio.Position(), a.audio.Duration()
	if dur > 0 {
		fraction := float64(pos) / float64(dur)
		if fraction < 0 {
			fraction = 0
		}

		if fraction > 1 {
			fraction = 1
		}

		a.view.Progress = int(fraction * percentScale)
		a.view.Elapsed = formatSeconds(int(pos.Seconds()))

		rem := int((dur - pos).Seconds())
		if rem < 0 {
			rem = 0
		}

		a.view.Remaining = formatSeconds(rem)
	}

	credit := a.audio.Credit()
	switch {
	case a.audio.Loading():
		a.view.Credit = "Loading free audio..."
	case credit.Title != "":
		a.view.Credit = credit.Title + " - " + credit.Creator + " (" + credit.License + ")"
	case a.audio.Err() != nil:
		a.view.Credit = "Audio unavailable"
	}
}

// setProgress moves the seek bar and recomputes both time strings.
func (a *App) setProgress(progress int) {
	if progress < 0 {
		progress = 0
	}

	if progress > progressMax {
		progress = 100
	}

	// Reuse the incoming progress value and only reformat the time strings
	// when the displayed progress actually changes.
	if progress != a.view.Progress {
		a.view.Progress = progress
		elapsed := a.seconds * progress / progressMax
		a.view.Elapsed = formatSeconds(elapsed)
		a.view.Remaining = formatSeconds(a.seconds - elapsed)
	} else {
		a.view.Progress = progress
	}

	a.page.Probe().Invalidate("seek")
	a.page.Probe().Invalidate("elapsed")
	a.page.Probe().Invalidate("remaining")
	a.logEvent("progress", "setProgress")
}

// resetProgress returns the seek bar to the start of the current track.
func (a *App) resetProgress() {
	oldSeconds := a.seconds
	a.seconds = lengthSeconds(a.view.Now.Length)
	a.view.Progress = 0
	a.view.Elapsed = "0:00"

	// Keep the formatted remaining time when the track length is unchanged.
	if oldSeconds != a.seconds || a.view.Remaining == "" {
		a.view.Remaining = formatSeconds(a.seconds)
	}

	a.page.Probe().Invalidate("seek")
	a.page.Probe().Invalidate("elapsed")
	a.page.Probe().Invalidate("remaining")
	a.logEvent("progress", "resetProgress")
}
