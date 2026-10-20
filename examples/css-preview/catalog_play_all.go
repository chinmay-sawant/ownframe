package main

import (
	"context"
	"time"
)

const playAllID = "play-all"

// playAll starts every preview the same way its own Start button does. The
// catalog stays windowed: the rows on screen start now, and the rows off
// screen wait for their first render, which gives the transition its
// starting paint.
func (d *demo) playAll(ctx context.Context) error {
	d.allPlaying = true
	d.playAllAt = time.Now()
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
		s.origin = d.playAllAt
		if s.base == nil {
			if box, ok := findBox(d.p.Boxes(), target.id); ok {
				s.base = capture(d.p.Display(), box)
			}
		}
		if s.base != nil {
			d.v.Started[target.key] = true
			d.v.Started[row.Button] = true
		}
		d.pending = append(d.pending, target.key)
	}
	d.v.Readout = "All previews are playing."
	d.p.SetData(d.v)
	if !walkActive() {
		d.p.SetTick(d.tick)
	}
	return nil
}
