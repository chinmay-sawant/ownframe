package main

import "time"

// begin binds one pending sample to the current display. It reports whether
// the sample started. A target that scrolled out of the window keeps its
// pending key so the animation starts when the row comes back.
func (d *demo) begin(key string) (bool, error) {
	t, ok := targetByKey(key)
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
	plan, err := bind(display, box, t, s.base, nil, s.origin)
	if err != nil {
		return false, err
	}
	s.plan = plan
	s.detached = false
	s.active = s.plan.duration > 0 && len(s.plan.pairs) > 0
	s.to = destination(s.plan.pairs)
	return true, nil
}

// destination copies the paints a sample moves toward. The first bind reads
// them from the display, where the started class already applied. Later
// binds reuse them, so a redraw in the middle of a transition does not make
// the destination the value the transition had reached.
func destination(pairs []pair) []paint {
	if len(pairs) == 0 {
		return nil
	}
	out := make([]paint, len(pairs))
	for i, p := range pairs {
		out[i] = p.to
	}
	return out
}
