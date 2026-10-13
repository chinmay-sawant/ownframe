package player

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

// ensureCredit keeps the now-bar credit line non-empty.
func (a *App) ensureCredit() {
	if a.view.Credit == "" {
		a.view.Credit = defaultCredit
	}
}

// animateCredit shows the loading note, or the clip behind the voice.
func (a *App) animateCredit(d *ownframe.Display, boxes []ownframe.Box) {
	text := ""

	switch {
	case a.audio.Loading():
		text = "Loading free audio..."
	case a.audio.Credit().Title != "":
		clip := a.audio.Credit()
		text = clip.Title + " - " + clip.Creator + " (" + clip.License + ")"
	case a.audio.Err() != nil:
		text = "Audio unavailable"
	}

	if text == "" {
		return
	}

	a.view.Credit = text

	box, ok := boxByID(boxes, "credit")
	if !ok {
		return
	}

	if op := frame.Text(d, box); op != nil && op.Text != text {
		op.Text = text
		a.page.Invalidate("credit")
	}
}

// updateView keeps the click-driven Redraw in step with the audio.
func (a *App) updateView(pos, dur time.Duration) {
	fraction := float64(pos) / float64(dur)
	if fraction < 0 {
		fraction = 0
	}

	if fraction > 1 {
		fraction = 1
	}

	a.view.Progress = int(fraction * 100)
	a.view.Elapsed = formatSeconds(int(pos.Seconds()))

	rem := int((dur - pos).Seconds())
	if rem < 0 {
		rem = 0
	}

	a.view.Remaining = formatSeconds(rem)
}
