package main

import "time"

// begin binds the pending sample to the current display. It reports whether
// the sample started. A target that scrolled out of the window keeps its
// pending key so the animation starts when the row comes back.
func (d *demo) begin() (bool, error) {
	t, ok := targetByKey(d.pendingKey)
	if !ok {
		return true, nil
	}
	s := d.samples[t.key]
	if s == nil {
		return true, nil
	}
	box, ok := findBox(d.p.Boxes(), t.id)
	if !ok {
		s.detached = true
		return false, nil
	}
	display := d.p.Display()
	s.origin = time.Now()
	plan, err := bind(display, box, t, s.base, s.origin)
	if err != nil {
		return false, err
	}
	s.plan = plan
	s.detached = false
	s.active = s.plan.duration > 0 && len(s.plan.pairs) > 0
	return true, nil
}
