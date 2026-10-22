package main

import "time"

// rebind rebuilds every sample against the current display. A sample whose
// row left the window keeps its origin, so the animation resumes where it
// was when the row comes back.
func (d *demo) rebind() error {
	now := time.Now()
	boxes := d.p.Boxes()
	for key, s := range d.samples {
		t, ok := targetByKey(key)
		if !ok {
			continue
		}
		boxIndex, ok := findBoxIndex(boxes, t.id)
		if !ok {
			// The row left the window. The sample keeps its origin and its
			// destination, and drops the old operations so the display they
			// point at can be released.
			s.detached = true
			s.plan = motion{}
			continue
		}
		s.detached = false
		box := boxes[boxIndex]
		plan, err := bind(d.p.Probe().Display(), box, t, s.base, s.to, s.origin)
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
