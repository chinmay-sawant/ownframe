package main

import (
	"context"
	"time"
)

const playAllID = "play-all"

// readoutID is the display-list readout the app fills from the template.
const readoutID = "preview-metadata"

// playAll starts every preview the same way its own Start button does: the
// class is a template value and one Redraw lays every started card out again.
func (d *demo) playAll(ctx context.Context) error {
	for _, row := range d.v.Catalog {
		if d.v.Started[row.Button] {
			d.v.Restart[row.Button] = !d.v.Restart[row.Button]
			d.v.Restarting[row.Button] = true
		}
		d.v.Started[row.Button] = true
		d.prepareManual(row)
	}
	d.v.Selected = ""
	d.v.Readout = "All previews are playing."
	d.readout = d.v.Readout
	d.startedAt = time.Now()
	d.startedSeq++
	return d.commit(ctx)
}
