package main

import (
	"image"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/chinmay-sawant/ownframe/internal/frame"
)

// stepSample advances one sample and marks where its operations were and
// where they landed, so the window repaints only those tiles.
func (d *demo) stepSample(s *sample, now time.Time) bool {
	display := d.p.Display()
	before := sampleBox(display, s.plan.pairs)
	active := s.plan.step(now)
	d.p.MarkRect(before.Union(sampleBox(display, s.plan.pairs)))
	return active
}

func sampleBox(display *ownframe.Display, pairs []pair) image.Rectangle {
	var box image.Rectangle
	for i := range pairs {
		box = box.Union(frame.OpBox(display, pairs[i].op))
	}
	return box
}

func (d *demo) rebind() error {
	now := time.Now()
	for key, s := range d.samples {
		t, ok := targetByKey(key)
		if !ok {
			continue
		}
		boxes := d.p.Boxes()
		boxIndex, ok := findBoxIndex(boxes, t.id)
		if !ok {
			continue
		}
		box := boxes[boxIndex]
		plan, err := bind(d.p.Display(), box, t, s.base, s.origin)
		if err != nil {
			if s.active {
				return err
			}
			continue
		}
		s.plan = plan
		if plan.animation != nil {
			if err := dispatchAnimationTriggers(boxes, boxIndex, s.source, plan.animation, s.origin); err != nil {
				return err
			}
		}
		if !s.active {
			s.plan.step(now)
		}
	}
	return nil
}
