package main

import (
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// begin binds one pending sample to the current display. It reports whether
// the sample started. The origin is the clock of the run that queued the
// sample, so a sample that starts now plays from its beginning. A target
// that scrolled out of the window is left to the scroll window, which
// queues it again when the row comes back.
func (d *demo) begin(key string, boxes []ownframe.Box, origin time.Time) (bool, error) {
	t, ok := targetByKey(key)
	if !ok {
		return true, nil
	}
	s := d.samples[t.key]
	if s == nil {
		return true, nil
	}
	box, ok := findBox(boxes, t.id)
	if !ok {
		// The row left the window. The scroll window queues the sample
		// again when the row comes back.
		s.detached = true
		return false, nil
	}
	s.detached = false
	if s.active {
		// The sample already runs. A second bind would read the value
		// the transition has reached as its destination.
		return true, nil
	}
	if s.base == nil {
		// The row has not rendered yet, so the starting paint is unknown.
		// The key stays pending until the row comes into view.
		return false, nil
	}
	if s.origin.IsZero() {
		s.origin = origin
	}
	plan, err := bind(d.p.Probe().Display(), box, t, s.base, nil, s.origin)
	if err != nil {
		return false, err
	}
	s.plan = plan
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
