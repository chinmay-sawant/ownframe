package player

// cardLength is the duration a card plays when it has no track length.
const cardLength = "3:30"

// setNow swaps the now-playing item for a card.
func (a *App) setNow(card Card) {
	a.view.Now = Track{Index: card.Index, Title: clip(card.Title, 30), Artist: clip(card.Sub, 30), Length: cardLength, Cover: card.Cover}
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

// setProgress moves the seek bar and recomputes both time strings.
func (a *App) setProgress(progress int) {
	if progress < 0 {
		progress = 0
	}

	if progress > 100 {
		progress = 100
	}

	a.view.Progress = progress
	elapsed := a.seconds * progress / 100
	a.view.Elapsed = formatSeconds(elapsed)
	a.view.Remaining = formatSeconds(a.seconds - elapsed)
}

// resetProgress returns the seek bar to the start of the current track.
func (a *App) resetProgress() {
	a.seconds = lengthSeconds(a.view.Now.Length)
	a.view.Progress = 0
	a.view.Elapsed = "0:00"
	a.view.Remaining = formatSeconds(a.seconds)
}
