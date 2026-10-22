package main

import (
	"context"
	"time"
)

const playAllID = "play-all"

// readoutID is the display-list snapshot the click patches.
const readoutID = "preview-metadata"

// playAll starts every preview the same way its own Start button does. The
// catalog stays windowed: the rows on screen start now, and the rows off
// screen wait for their first render, which gives the transition its
// starting paint.
//
// The run keeps no clock of its own. Every sample takes its start time from
// the place a single Start takes it, when the sample binds to the display.
// A clock stamped here would be spent by the redraw that follows the click,
// and the shortest sample would be over before it began.
func (d *demo) playAll(ctx context.Context) error {
	d.allPlaying = true
	d.playAllAt = time.Time{}
	for _, row := range d.v.Catalog {
		target, ok := targetByKey(row.SampleKey)
		if !ok {
			continue
		}
		s := d.samples[target.key]
		if s == nil {
			s = &sample{}
			d.samples[target.key] = s
		}
		s.source = row.Button
		if s.base == nil {
			if box, ok := findBox(d.p.Boxes(), target.id); ok {
				s.base = capture(d.p.Display(), box)
			}
		}
		d.v.Started[target.key] = true
		d.v.Started[row.Button] = true
		// The sample takes its start time when it binds, so a run that
		// starts now plays from the beginning.
		s.origin = time.Time{}
		s.active = false
		if s.base == nil {
			// The row is off screen. The scroll window queues it when
			// the row comes into view.
			continue
		}
		d.queue(target.key)
	}
	// The card's class carries the started state into the tree, so the
	// sample's target moves on the next layout.
	if !d.patchGrid(ctx) {
		// The engine refused the grid, so the classes land in the tree
		// and the next layout shows them.
		d.treeDirty = true
	}
	d.setReadout(ctx, "All previews are playing.")
	d.p.SetData(d.v)
	if !walkActive() {
		d.p.SetTick(d.tick)
	}
	return nil
}
