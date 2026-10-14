package player

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

// animateSeek stretches the accent fill to the playing fraction.
func (a *App) animateSeek(d *ownframe.Display, boxes []ownframe.Box, pos, dur time.Duration) {
	seek, ok := boxByID(boxes, "seek")
	if !ok {
		return
	}

	fill := frame.Fill(d, seek, accent)
	if fill == nil {
		return
	}

	_, _, w, _ := frame.BoxUnits(d, seek)
	fraction := float64(pos) / float64(dur)
	if fraction < 0 {
		fraction = 0
	}

	if fraction > 1 {
		fraction = 1
	}

	oldW := fill.W
	fill.W = w * fraction

	if fill.W != oldW {
		a.page.Invalidate("seek")
	}
}

// animateTimes writes the elapsed and remaining labels when they change.
func (a *App) animateTimes(d *ownframe.Display, boxes []ownframe.Box, pos, dur time.Duration) {
	if box, ok := boxByID(boxes, "elapsed"); ok {
		if op := frame.Text(d, box); op != nil {
			if text := formatSeconds(int(pos.Seconds())); op.Text != text {
				op.Text = text
				a.page.Invalidate("elapsed")
			}
		}
	}

	if box, ok := boxByID(boxes, "remaining"); ok {
		if op := frame.Text(d, box); op != nil {
			rem := int((dur - pos).Seconds())
			if rem < 0 {
				rem = 0
			}

			if text := "-" + formatSeconds(rem); op.Text != text {
				op.Text = text
				a.page.Invalidate("remaining")
			}
		}
	}
}
