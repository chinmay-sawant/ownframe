package ui

import (
	"image"
	"strconv"

	"github.com/chinmay-sawant/ownframe"
)

// rowStrip is the full-width line of one active row, in CSS pixels. The
// strip has no element id, so it cannot steal a click from a button.
func rowStrip(byID map[string]ownframe.Box, i int, width, _ int) image.Rectangle {
	var top, bot float64
	seen := false
	key := strconv.Itoa(i)

	for _, id := range []string{"track-" + key, "pct-" + key, "spd-" + key, "eta-" + key} {
		box, ok := byID[id]
		if !ok {
			continue
		}

		if !seen || box.Y < top {
			top = box.Y
		}

		if !seen || box.Y+box.H > bot {
			bot = box.Y + box.H
		}

		seen = true
	}

	if !seen {
		return image.Rectangle{}
	}

	if width < 1 {
		width = int(bot)
	}

	return image.Rect(0, int(top)-stripPad, width, int(bot)+stripPad)
}
