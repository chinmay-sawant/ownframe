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
func (d *demo) playAll(ctx context.Context) error {
	d.allPlaying = true
	started := make([]string, 0, len(d.v.Catalog))
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
		if s.base != nil {
			d.v.Started[target.key] = true
			d.v.Started[row.Button] = true
			// The card's class carries the started state into the tree, so
			// the sample's target moves on the next layout.
			d.patchCard(ctx, row)
		}
		started = append(started, target.key)
		d.pending = append(d.pending, target.key)
	}
	// The shared start time is stamped after the patching above. A click that
	// patches every card can take longer than the shortest animation, and a
	// clock that starts before that work leaves every sample finished before
	// the first tick reaches it.
	d.playAllAt = time.Now()
	for _, key := range started {
		if s := d.samples[key]; s != nil {
			s.origin = d.playAllAt
		}
	}
	d.v.Readout = "All previews are playing."
	d.p.SetData(d.v)
	if !walkActive() {
		d.p.SetTick(d.tick)
	}
	return nil
}
