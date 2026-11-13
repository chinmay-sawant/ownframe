package player

import (
	"github.com/chinmay-sawant/ownframe"
)

// ensureCredit keeps the now-bar credit line non-empty.
func (a *App) ensureCredit() {
	if a.view.Credit == "" {
		a.view.Credit = defaultCredit
	}
}

// animateCredit shows the loading note, or the clip behind the voice. It
// patches only the retained credit op; template truth moves at commit time
// through syncViewFromAudio, never from the tick.
func (a *App) animateCredit(d *ownframe.Display, boxes []ownframe.Box) {
	text := ""
	credit := a.audio.Credit()

	switch {
	case a.audio.Loading():
		text = "Loading free audio..."
	case credit.Title != "":
		text = credit.Title + " - " + credit.Creator + " (" + credit.License + ")"
	case a.audio.Err() != nil:
		text = "Audio unavailable"
	}

	if text == "" {
		return
	}

	box, ok := boxByID(boxes, "credit")
	if !ok {
		return
	}

	if op := ownframe.Text(d, box); op != nil && op.Text != text {
		op.Text = text
		a.page.Probe().Invalidate("credit")
	}
}
