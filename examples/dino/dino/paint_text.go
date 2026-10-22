package dino

import (
	"fmt"

	"github.com/chinmay-sawant/ownframe"
)

// ink is #535353 and muted is #9a9a9a, as 0..1 channels.
var (
	ink   = [3]float64{83 / 255.0, 83 / 255.0, 83 / 255.0}
	muted = [3]float64{154 / 255.0, 154 / 255.0, 154 / 255.0}
)

// paintText refreshes the score, the best, the frame rate, and the two
// overlay messages.
func (a *App) paintText() {
	setText(a.parts.score, fmt.Sprintf("HI %05d %05d", a.game.high, a.game.score))

	if a.fps.fps > 0 {
		setText(a.parts.fps, fmt.Sprintf("%03d FPS", min(a.fps.fps, maxFPS)))
	}

	start, keys, again := startText, keysText, againText
	if a.touch {
		start, keys, again = tapStartText, tapKeysText, tapAgainText
	}

	revealText(a.parts.start, a.game.phase == ready, start, ink)
	revealText(a.parts.keys, a.game.phase == ready, keys, muted)
	revealText(a.parts.over, a.game.phase == over, overText, ink)
	revealText(a.parts.again, a.game.phase == over, again, muted)
}

// setText changes a text operation only when the string changed.
func setText(op *ownframe.DisplayOp, text string) {
	if op != nil && op.Text != text {
		op.Text = text
	}
}

// revealText shows a message in its colour, or empties it so the replay
// draws nothing. An empty run is how a text operation hides: the replay
// reads opacity, not the Alpha paint field.
func revealText(op *ownframe.DisplayOp, visible bool, text string, color [3]float64) {
	if op == nil {
		return
	}

	if !visible {
		setText(op, "")

		return
	}

	op.R, op.G, op.B = color[0], color[1], color[2]
	setText(op, text)
}
