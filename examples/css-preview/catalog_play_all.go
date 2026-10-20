package main

import (
	"context"
)

const playAllID = "play-all"

// playAll starts every preview the same way its own Start button does. The
// whole catalog renders first, so every preview has a box and a starting
// paint, then each one goes through the same start path a click uses.
func (d *demo) playAll(ctx context.Context) error {
	d.v.Rows = d.v.Catalog
	d.v.TopPad, d.v.BotPad = 0, 0
	d.allPlaying = true
	d.p.SetData(d.v)
	if err := d.p.Redraw(ctx); err != nil {
		return err
	}

	for _, row := range d.v.Catalog {
		target, ok := targetByKey(row.SampleKey)
		if !ok {
			continue
		}
		if err := d.start(target, row.Button); err != nil {
			return err
		}
	}
	d.v.Readout = "All previews are playing."

	return nil
}
