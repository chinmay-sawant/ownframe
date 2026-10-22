package player

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

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
