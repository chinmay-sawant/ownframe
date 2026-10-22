package main

import (
	"image"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

// stepSamples advances every sample that is on screen. It reports whether a
// sample is still running and whether one is waiting for its row to come
// back into the window.
func (d *demo) stepSamples(now time.Time, partial bool) (live, waiting bool) {
	for _, s := range d.samples {
		if !s.active {
			continue
		}
		if s.detached {
			waiting = true
			continue
		}
		var running bool
		if partial {
			running = d.stepSample(s, now)
		} else {
			running = s.plan.step(now)
		}
		if !running {
			s.active = false
		}
		live = s.active || live
	}
	return live, waiting
}

// stepSample advances one sample and marks where its operations were and
// where they landed, so the window repaints only those tiles.
func (d *demo) stepSample(s *sample, now time.Time) bool {
	display := d.p.Testing().Display()
	before := sampleBox(display, s.plan.pairs)
	active := s.plan.step(now)
	d.p.Testing().MarkRect(before.Union(sampleBox(display, s.plan.pairs)))
	return active
}

func sampleBox(display *ownframe.Display, pairs []pair) image.Rectangle {
	var box image.Rectangle
	for i := range pairs {
		box = box.Union(frame.OpBox(display, pairs[i].op))
	}
	return box
}
